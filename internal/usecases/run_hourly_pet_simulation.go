package usecases

import (
	"hash/fnv"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/timeutil"
)

type RunHourlyPetSimulationInput struct {
	SimulatedAt *time.Time
}

type RunHourlyPetSimulationOutput struct {
	SimulatedAt time.Time                         `json:"simulated_at"`
	TotalPets   int                               `json:"total_pets"`
	Processed   int                               `json:"processed"`
	Skipped     int                               `json:"skipped"`
	Results     []RunHourlyPetSimulationPetResult `json:"results"`
}

type RunHourlyPetSimulationPetResult struct {
	PetID           string  `json:"pet_id"`
	PreviousGroupID *int    `json:"previous_group_master_id,omitempty"`
	NextGroupID     int     `json:"next_group_master_id"`
	Moved           bool    `json:"moved"`
	MoveProbability float64 `json:"move_probability"`
	AmbientEvent    string  `json:"ambient_event"`
}

type RunHourlyPetSimulation struct {
	repo domain.PetSimulationRepository
}

const (
	hourlySouvenirDropRate  = 0.05
	groupDeltaRange         = 0.14
	maxIntentCategoryPool   = 4
	maxInterestScoreBonus   = 0.18
	interestScoreScale      = 3.0
	interestSelectionWeight = 12.0
)

func NewRunHourlyPetSimulation(repo domain.PetSimulationRepository) *RunHourlyPetSimulation {
	return &RunHourlyPetSimulation{repo: repo}
}

func (u *RunHourlyPetSimulation) Execute(input RunHourlyPetSimulationInput) (RunHourlyPetSimulationOutput, error) {
	// 1時間単位に成形
	simulatedAt := floorToHour(timeutil.NowJST())
	if input.SimulatedAt != nil {
		simulatedAt = floorToHour(input.SimulatedAt.In(timeutil.LocationJST()))
	}

	pets, err := u.repo.FindActivePetsForSimulation()
	if err != nil {
		return RunHourlyPetSimulationOutput{}, err
	}

	groups, err := u.repo.FindActiveGroupsForSimulation()
	if err != nil {
		return RunHourlyPetSimulationOutput{}, err
	}
	if len(groups) == 0 {
		return RunHourlyPetSimulationOutput{}, domain.ErrValidation
	}
	groupInterests, err := u.repo.FindGroupInterestsForSimulation()
	if err != nil {
		return RunHourlyPetSimulationOutput{}, err
	}

	output := RunHourlyPetSimulationOutput{
		SimulatedAt: simulatedAt,
		TotalPets:   len(pets),
		Results:     make([]RunHourlyPetSimulationPetResult, 0, len(pets)),
	}

	for _, pet := range pets {
		// 行動計画の作成
		plan := u.planPetHour(pet, groups, groupInterests[pet.ID()], simulatedAt)
		// db保存
		saved, err := u.repo.SaveHourlySimulation(plan.saveInput)
		if err != nil {
			return RunHourlyPetSimulationOutput{}, err
		}
		if !saved {
			output.Skipped++
			continue
		}

		output.Processed++
		output.Results = append(output.Results, plan.result)
	}

	return output, nil
}

type petHourPlan struct {
	saveInput domain.PetSimulationSaveInput
	result    RunHourlyPetSimulationPetResult
}

// 行動計画の作成
func (u *RunHourlyPetSimulation) planPetHour(
	pet domain.SimulationPet,
	groups []domain.GroupMaster,
	interests domain.GroupInterestScores,
	simulatedAt time.Time,
) petHourPlan {
	currentGroupID := pet.CurrentGroupMasterID()
	currentGroup := findGroup(groups, currentGroupID)
	if currentGroup == nil {
		currentGroup = &groups[0]
	}

	r := rand.New(rand.NewSource(simulationSeed(string(pet.ID()), simulatedAt)))
	// 移動確率の計算
	metrics := calculateSimulationMetrics(pet, *currentGroup, simulatedAt)
	moved := r.Float64() < metrics.moveProbability
	// 次の群れを選択する
	nextGroup := *currentGroup
	if moved || currentGroupID == nil {
		nextGroup = chooseNextGroup(pet, groups, currentGroup.ID(), metrics.restNeed, interests, r)
		moved = currentGroupID == nil || nextGroup.ID() != currentGroup.ID()
	}

	interactionCount := calculateInteractionCount(pet, nextGroup, r)
	souvenirDrop := shouldDropSouvenir(r)
	souvenirNote := buildSouvenirNote(nextGroup)
	ambientEvent, reportMaterial := buildAmbientText(nextGroup, moved, metrics.restNeed, interactionCount, r)
	// hourly log
	log := domain.NewPetHourlyLog(
		domain.NewPetHourlyLogID(),
		pet.ID(),
		nextGroup.ID(),
		pet.CurrentJoinID,
		simulatedAt,
		!moved,
		metrics.moveProbability,
		metrics.boredom,
		metrics.restNeed,
		metrics.currentGroupFit,
		metrics.attachmentToCurrentGroup,
		metrics.recentMovePenalty,
		nextGroup.EnergyDelta(),
		nextGroup.CuriosityDelta(),
		nextGroup.SocialityDelta(),
		nextGroup.RoutineDelta(),
		interactionCount,
		&ambientEvent,
		&reportMaterial,
		timeutil.NowJST(),
	)

	var previousGroupID *domain.GroupMasterID
	var previousGroupIDForResult *int
	if currentGroupID != nil {
		value := domain.GroupMasterID(*currentGroupID)
		previousGroupID = &value
		resultValue := int(value)
		previousGroupIDForResult = &resultValue
	}

	return petHourPlan{
		// repositoryに渡すデータを作成
		saveInput: domain.PetSimulationSaveInput{
			PetID:           pet.ID(),
			PreviousGroupID: previousGroupID,
			NextGroupID:     nextGroup.ID(),
			PreviousJoinID:  pet.CurrentJoinID,
			MoveReason:      "hourly_simulation",
			Moved:           moved,
			EnergyDelta:     nextGroup.EnergyDelta(),
			CuriosityDelta:  nextGroup.CuriosityDelta(),
			SocialityDelta:  nextGroup.SocialityDelta(),
			RoutineDelta:    nextGroup.RoutineDelta(),
			SimulatedAt:     simulatedAt,
			Log:             log,
			SouvenirDrop:    souvenirDrop,
			SouvenirNote:    souvenirNote,
		},
		result: RunHourlyPetSimulationPetResult{
			PetID:           string(pet.ID()),
			PreviousGroupID: previousGroupIDForResult,
			NextGroupID:     int(nextGroup.ID()),
			Moved:           moved,
			MoveProbability: metrics.moveProbability,
			AmbientEvent:    ambientEvent,
		},
	}
}

type simulationMetrics struct {
	moveProbability          float64
	boredom                  float64
	restNeed                 float64
	currentGroupFit          float64
	attachmentToCurrentGroup float64
	recentMovePenalty        float64
}

// 移動確率の計算
func calculateSimulationMetrics(pet domain.SimulationPet, group domain.GroupMaster, simulatedAt time.Time) simulationMetrics {
	hoursInGroup := 0.0
	if pet.JoinedAt != nil {
		hoursInGroup = math.Max(0, simulatedAt.Sub(*pet.JoinedAt).Hours())
	}

	boredom := clamp(hoursInGroup/24, 0, 0.35)
	restNeed := clamp((45-pet.Energy())/45, 0, 1)
	currentGroupFit := calculateGroupFit(pet.Pet, group)
	attachment := clamp(pet.Sociality()/500+hoursInGroup/120, 0, 0.25)
	recentMovePenalty := 0.0
	if pet.JoinedAt != nil {
		elapsedSinceJoin := simulatedAt.Sub(*pet.JoinedAt)
		if elapsedSinceJoin >= 0 && elapsedSinceJoin < 3*time.Hour {
			recentMovePenalty = 0.08
		}
	}
	curiosityBonus := clamp((pet.Curiosity()-50)/250, -0.08, 0.20)

	moveProbability := clamp(
		0.34+boredom+curiosityBonus+(1-currentGroupFit)*0.15+restNeed*0.25-attachment-recentMovePenalty,
		0.02,
		0.85,
	)

	return simulationMetrics{
		moveProbability:          moveProbability,
		boredom:                  boredom,
		restNeed:                 restNeed,
		currentGroupFit:          currentGroupFit,
		attachmentToCurrentGroup: attachment,
		recentMovePenalty:        recentMovePenalty,
	}
}

// 群れとの相性計算
func calculateGroupFit(pet domain.Pet, group domain.GroupMaster) float64 {
	score := 0.5
	score += needWeight(100-pet.Energy(), group.EnergyDelta())
	score += needWeight(pet.Curiosity(), group.CuriosityDelta())
	score += needWeight(pet.Sociality(), group.SocialityDelta())
	score += needWeight(100-pet.Routine(), -group.RoutineDelta())
	return clamp(score, 0, 1)
}

func needWeight(status float64, delta float64) float64 {
	return clamp(status/100, 0, 1) * delta * 1.5
}

type nextGroupCandidate struct {
	group         domain.GroupMaster
	score         float64
	interestBonus float64
}

type movementIntent struct {
	name               string
	categories         []string
	categoryPoolSize   int
	energyNeed         float64
	curiosityNeed      float64
	socialityNeed      float64
	routineNeed        float64
	energyWeight       float64
	curiosityWeight    float64
	socialityWeight    float64
	routineWeight      float64
	protectEnergy      bool
	protectRoutine     bool
	allowNoisyDrift    bool
	categoryScoreBonus float64
}

// 移動先の群れを選択
func chooseNextGroup(
	pet domain.SimulationPet,
	groups []domain.GroupMaster,
	currentID domain.GroupMasterID,
	restNeed float64,
	interests domain.GroupInterestScores,
	r *rand.Rand,
) domain.GroupMaster {
	intent := buildMovementIntent(pet, restNeed)
	candidates := make([]nextGroupCandidate, 0, len(groups))
	for _, group := range groups {
		if group.ID() == currentID && len(groups) > 1 {
			continue
		}

		interestBonus := groupInterestBonus(interests[group.ID()])
		score := calculateIntentGroupScore(pet.Pet, group, intent) + interestBonus
		candidates = append(candidates, nextGroupCandidate{
			group:         group,
			score:         score,
			interestBonus: interestBonus,
		})
	}

	if len(candidates) == 0 {
		return groups[0]
	}

	categoryCandidates := bestCandidateByCategory(candidates)
	sort.SliceStable(categoryCandidates, func(i, j int) bool {
		return categoryCandidates[i].score > categoryCandidates[j].score
	})

	categoryPoolSize := intent.categoryPoolSize
	if categoryPoolSize < 1 {
		categoryPoolSize = 1
	}
	if categoryPoolSize > maxIntentCategoryPool {
		categoryPoolSize = maxIntentCategoryPool
	}
	if categoryPoolSize > len(categoryCandidates) {
		categoryPoolSize = len(categoryCandidates)
	}

	selectedCategory := groupCategory(
		chooseWeightedCandidate(categoryCandidates[:categoryPoolSize], r).group,
	)
	categoryGroups := candidatesForCategory(candidates, selectedCategory)
	sort.SliceStable(categoryGroups, func(i, j int) bool {
		return categoryGroups[i].score > categoryGroups[j].score
	})

	groupPoolSize := closeCandidatePoolSize(categoryGroups)
	return chooseWeightedCandidate(categoryGroups[:groupPoolSize], r).group
}

// 興味スコアは投稿ごとに累積するため、加点は飽和させてステータス由来の行動を優先する。
func groupInterestBonus(interestScore float64) float64 {
	if interestScore <= 0 {
		return 0
	}
	return maxInterestScoreBonus * (1 - math.Exp(-interestScore/interestScoreScale))
}

// 興味がない候補は従来どおり均等に扱い、興味を持つ群れだけ抽選確率を上げる。
func chooseWeightedCandidate(candidates []nextGroupCandidate, r *rand.Rand) nextGroupCandidate {
	if len(candidates) == 1 {
		return candidates[0]
	}

	totalWeight := 0.0
	for _, candidate := range candidates {
		totalWeight += candidateSelectionWeight(candidate)
	}

	target := r.Float64() * totalWeight
	for _, candidate := range candidates {
		target -= candidateSelectionWeight(candidate)
		if target <= 0 {
			return candidate
		}
	}

	return candidates[len(candidates)-1]
}

func candidateSelectionWeight(candidate nextGroupCandidate) float64 {
	return 1 + candidate.interestBonus*interestSelectionWeight
}

func buildMovementIntent(pet domain.SimulationPet, restNeed float64) movementIntent {
	base := movementIntent{
		name:               "balanced",
		categories:         []string{"life", "hobby", "creation", "work_study", "digital", "special", "thinking", "condition", "place"},
		categoryPoolSize:   3,
		energyNeed:         clamp((50-pet.Energy())/50, -0.7, 1),
		curiosityNeed:      clamp((pet.Curiosity()-50)/50, -0.5, 1),
		socialityNeed:      clamp((pet.Sociality()-50)/50, -0.5, 1),
		routineNeed:        clamp((55-pet.Routine())/55, -0.4, 1),
		energyWeight:       1.0,
		curiosityWeight:    1.0,
		socialityWeight:    1.0,
		routineWeight:      1.0,
		categoryScoreBonus: 0.12,
	}

	switch {
	case restNeed > 0.55 || pet.Energy() < 30:
		base.name = "rest"
		base.categories = []string{"life", "special", "place", "condition", "thinking"}
		base.categoryPoolSize = 2
		base.energyNeed = 1
		base.curiosityNeed = -0.35
		base.socialityNeed = -0.25
		base.routineNeed = 0.8
		base.energyWeight = 1.8
		base.routineWeight = 1.5
		base.protectEnergy = true
		base.protectRoutine = true
	case pet.Routine() < 35:
		base.name = "routine_recovery"
		base.categories = []string{"life", "place", "work_study", "hobby", "thinking"}
		base.categoryPoolSize = 2
		base.energyNeed = 0.45
		base.curiosityNeed = 0.15
		base.socialityNeed = 0.1
		base.routineNeed = 1
		base.routineWeight = 1.9
		base.protectRoutine = true
	case pet.Energy() > 78:
		base.name = "active"
		base.categories = []string{"condition", "place", "hobby", "life", "special"}
		base.categoryPoolSize = 3
		base.energyNeed = -0.9
		base.curiosityNeed = clamp((pet.Curiosity()-45)/55, 0, 1)
		base.socialityNeed = clamp((pet.Sociality()-45)/55, 0, 1)
		base.routineNeed = clamp((55-pet.Routine())/55, -0.2, 0.6)
		base.energyWeight = 1.4
		base.allowNoisyDrift = pet.Routine() > 60
	case pet.Curiosity() > 72:
		base.name = "explore"
		base.categories = []string{"creation", "hobby", "digital", "thinking", "place", "work_study", "special"}
		base.categoryPoolSize = 4
		base.energyNeed = clamp((45-pet.Energy())/45, -0.3, 0.7)
		base.curiosityNeed = 1
		base.socialityNeed = clamp((pet.Sociality()-45)/55, -0.1, 0.8)
		base.routineNeed = clamp((50-pet.Routine())/50, -0.2, 0.8)
		base.curiosityWeight = 1.5
		base.protectEnergy = pet.Energy() < 45
		base.protectRoutine = pet.Routine() < 45
	case pet.Sociality() > 72:
		base.name = "social"
		base.categories = []string{"hobby", "life", "digital", "special", "condition", "creation", "work_study"}
		base.categoryPoolSize = 3
		base.energyNeed = clamp((45-pet.Energy())/45, -0.2, 0.8)
		base.curiosityNeed = clamp((pet.Curiosity()-45)/55, -0.1, 0.8)
		base.socialityNeed = 1
		base.routineNeed = clamp((50-pet.Routine())/50, -0.2, 0.8)
		base.socialityWeight = 1.5
		base.protectEnergy = pet.Energy() < 45
		base.protectRoutine = pet.Routine() < 45
		base.allowNoisyDrift = pet.Energy() > 60 && pet.Routine() > 60
	}

	return base
}

// 移動意図と群れdeltaの噛み合いを点数化す
func calculateIntentGroupScore(pet domain.Pet, group domain.GroupMaster, intent movementIntent) float64 {
	score := 0.45
	score += directionalDeltaScore(intent.energyNeed, group.EnergyDelta(), intent.energyWeight)
	score += directionalDeltaScore(intent.curiosityNeed, group.CuriosityDelta(), intent.curiosityWeight)
	score += directionalDeltaScore(intent.socialityNeed, group.SocialityDelta(), intent.socialityWeight)
	score += directionalDeltaScore(intent.routineNeed, group.RoutineDelta(), intent.routineWeight)
	score += categoryPriorityBonus(group, intent)

	if intent.protectEnergy && group.EnergyDelta() < 0 {
		score -= math.Abs(group.EnergyDelta()) / groupDeltaRange * 0.55
	}
	if intent.protectRoutine && group.RoutineDelta() < 0 {
		score -= math.Abs(group.RoutineDelta()) / groupDeltaRange * 0.65
	}
	if !intent.allowNoisyDrift && isNoisyDriftGroup(group) {
		score -= noisyDriftPenalty(pet, group)
	}

	return score
}

// 「上げたい/下げたいステータス」と群れdeltaの向きを比較
// needが正ならdeltaが正の群れ、needが負ならdeltaが負の群れを高く評価
func directionalDeltaScore(need float64, delta float64, weight float64) float64 {
	if math.Abs(need) < 0.05 {
		return -math.Abs(delta) / groupDeltaRange * 0.04 * weight
	}
	return need * clamp(delta/groupDeltaRange, -1, 1) * 0.28 * weight
}

// 移動意図に合うカテゴリを少し優先
// deltaだけで決めると似た強さの群れに偏るため、カテゴリで行き先の意味を補強
func categoryPriorityBonus(group domain.GroupMaster, intent movementIntent) float64 {
	category := groupCategory(group)
	for i, priority := range intent.categories {
		if category == priority {
			rank := float64(len(intent.categories)-i) / float64(len(intent.categories))
			return rank * intent.categoryScoreBonus
		}
	}
	return -0.10
}

// 楽しいが体力と生活リズムを削りやすい群れを判定する
// 低体力・低routineのpetが祭り/お酒/夜ふかし系に吸われすぎるのを防ぐために使用
func isNoisyDriftGroup(group domain.GroupMaster) bool {
	return group.EnergyDelta() < -0.08 &&
		group.RoutineDelta() < -0.07 &&
		(group.CuriosityDelta() > 0.10 || group.SocialityDelta() > 0.10)
}

// petの余裕が少ないほど騒がしい群れの点数を下げる
// 完全禁止ではなく、energy/routineが安定している時は行ける余地を残す
func noisyDriftPenalty(pet domain.Pet, group domain.GroupMaster) float64 {
	energyRisk := clamp((60-pet.Energy())/60, 0, 1)
	routineRisk := clamp((65-pet.Routine())/65, 0, 1)
	noisyDelta := (math.Abs(group.EnergyDelta()) + math.Abs(group.RoutineDelta())) / (groupDeltaRange * 2)
	return (0.20 + energyRisk*0.25 + routineRisk*0.35) * noisyDelta
}

// カテゴリごとの代表候補を1つずつ作成
// まずカテゴリを選んでから群れを選ぶことで、同じ系統の群ればかりに偏るのを抑える
func bestCandidateByCategory(candidates []nextGroupCandidate) []nextGroupCandidate {
	bestByCategory := make(map[string]nextGroupCandidate)
	for _, candidate := range candidates {
		category := groupCategory(candidate.group)
		best, ok := bestByCategory[category]
		if !ok || candidate.score > best.score {
			bestByCategory[category] = candidate
		}
	}

	grouped := make([]nextGroupCandidate, 0, len(bestByCategory))
	for _, candidate := range bestByCategory {
		grouped = append(grouped, candidate)
	}
	return grouped
}

// 選ばれたカテゴリ内の群れだけを取り出す
// カテゴリ決定後に、そのカテゴリ内でpetに一番合う群れを選ぶために使用
func candidatesForCategory(candidates []nextGroupCandidate, category string) []nextGroupCandidate {
	filtered := make([]nextGroupCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if groupCategory(candidate.group) == category {
			filtered = append(filtered, candidate)
		}
	}
	return filtered
}

// カテゴリ内で僅差の群れだけを候補に残す。
// ほぼ同点なら少し揺らぎを許し、差が大きい時は最適な群れを選びやすくする
func closeCandidatePoolSize(candidates []nextGroupCandidate) int {
	if len(candidates) <= 1 {
		return len(candidates)
	}

	poolSize := 1
	bestScore := candidates[0].score
	for i := 1; i < len(candidates) && i < 3; i++ {
		if bestScore-candidates[i].score > 0.08 {
			break
		}
		poolSize = i + 1
	}
	return poolSize
}

// カテゴリ未設定の群れにも安全にカテゴリ名を与える。
func groupCategory(group domain.GroupMaster) string {
	if group.Category() == nil || *group.Category() == "" {
		return "uncategorized"
	}
	return *group.Category()
}

// 交流会数を決める
func calculateInteractionCount(pet domain.SimulationPet, group domain.GroupMaster, r *rand.Rand) int {
	base := clamp(pet.Sociality()/100, 0, 1)
	if group.SocialityDelta() > 0 {
		base += group.SocialityDelta()
	}
	if r.Float64() < base*0.45 {
		return 1
	}
	return 0
}

func shouldDropSouvenir(r *rand.Rand) bool {
	return r.Float64() < hourlySouvenirDropRate
}

func buildSouvenirNote(group domain.GroupMaster) string {
	return group.DisplayName() + "で、小さなおみやげを見つけたようです。"
}

// reportLine は、ログ用の短いイベント名とレポート本文をまとめる。
type reportLine struct {
	event    string
	material string
}

// レポート用の文章を作成
// 投稿本文や他ユーザーの内容は出さず、群れ・移動状態・気配だけで表現を変える。
func buildAmbientText(group domain.GroupMaster, moved bool, restNeed float64, interactionCount int, r *rand.Rand) (string, string) {
	if restNeed > 0.65 {
		return chooseReportLine(r, []reportLine{
			{
				event:    "少し休みたそうにしていた",
				material: group.DisplayName() + "で、ペットは少し休みたそうに丸まっていました。",
			},
			{
				event:    "静かな場所で息を整えた",
				material: group.DisplayName() + "のすみで、ペットはゆっくり息を整えていました。",
			},
			{
				event:    "眠そうに過ごした",
				material: group.DisplayName() + "で、ペットは眠そうに目を細めていました。",
			},
			{
				event:    "やわらかい気配に包まれた",
				material: group.DisplayName() + "の近くで、やわらかい気配に包まれていました。",
			},
		})
	}
	if interactionCount > 0 {
		return chooseReportLine(r, []reportLine{
			{
				event:    "近くの気配と遊んだ",
				material: group.DisplayName() + "で、近くのYoYoと少しだけ遊んでいました。",
			},
			{
				event:    "となりの気配と並んだ",
				material: group.DisplayName() + "で、となりの気配としばらく並んでいました。",
			},
			{
				event:    "小さな輪に混ざった",
				material: group.DisplayName() + "で、小さな気配の輪にそっと混ざっていました。",
			},
			{
				event:    "気配を分け合った",
				material: group.DisplayName() + "で、近くの気配と小さなきらめきを分け合っていました。",
			},
		})
	}
	if moved {
		return chooseReportLine(r, movedReportLines(group))
	}
	return chooseReportLine(r, stayedReportLines(group))
}

func chooseReportLine(r *rand.Rand, lines []reportLine) (string, string) {
	if len(lines) == 0 {
		return "群れで過ごした", "ペットは静かに時間を過ごしていました。"
	}
	line := lines[r.Intn(len(lines))]
	return line.event, line.material
}

func movedReportLines(group domain.GroupMaster) []reportLine {
	groupName := group.DisplayName()
	switch groupCategory(group) {
	case "life":
		return []reportLine{
			{event: "暮らしの気配へ向かった", material: groupName + "へ向かい、暮らしの気配を確かめていました。"},
			{event: "いつもの気配を探した", material: groupName + "で、いつもの気配を少しずつ探していました。"},
			{event: "生活の音に近づいた", material: groupName + "のあたりで、生活の音に耳を澄ませていました。"},
		}
	case "hobby":
		return []reportLine{
			{event: "楽しそうな気配に近づいた", material: groupName + "へ向かい、楽しそうな気配を眺めていました。"},
			{event: "好きなものの匂いを追った", material: groupName + "で、好きなものの匂いを追いかけていました。"},
			{event: "遊びの気配を見つけた", material: groupName + "の近くで、遊びの気配を見つけたようです。"},
		}
	case "creation":
		return []reportLine{
			{event: "ものづくりの気配を眺めた", material: groupName + "へ向かい、ものづくりの気配をじっと眺めていました。"},
			{event: "手を動かす気配に寄った", material: groupName + "で、手を動かす気配のそばにいました。"},
			{event: "新しい形を探した", material: groupName + "のあたりで、新しい形を探していました。"},
		}
	case "work_study":
		return []reportLine{
			{event: "学びの気配を拾った", material: groupName + "へ向かい、学びの気配を少し拾ってきました。"},
			{event: "集中した空気に触れた", material: groupName + "で、集中した空気にそっと触れていました。"},
			{event: "考える気配に近づいた", material: groupName + "の近くで、考える気配に近づいていました。"},
		}
	case "digital":
		return []reportLine{
			{event: "明るい画面の気配を追った", material: groupName + "へ向かい、明るい画面の気配を追っていました。"},
			{event: "電子のきらめきを見た", material: groupName + "で、電子のきらめきを見上げていました。"},
			{event: "小さな通信の気配に触れた", material: groupName + "の近くで、小さな通信の気配に触れていました。"},
		}
	case "special":
		return []reportLine{
			{event: "にぎやかな気配へ向かった", material: groupName + "へ向かい、にぎやかな気配を遠くから眺めていました。"},
			{event: "特別な空気を確かめた", material: groupName + "で、いつもと違う空気を確かめていました。"},
			{event: "きらめく気配に近づいた", material: groupName + "の近くで、きらめく気配に近づいていました。"},
		}
	case "thinking":
		return []reportLine{
			{event: "考えごとの気配に寄り添った", material: groupName + "へ向かい、考えごとの気配に寄り添っていました。"},
			{event: "静かな問いを見つけた", material: groupName + "で、静かな問いを見つけたようです。"},
			{event: "遠くを見る気配に触れた", material: groupName + "のあたりで、遠くを見る気配に触れていました。"},
		}
	case "condition":
		return []reportLine{
			{event: "からだの気配を確かめた", material: groupName + "へ向かい、からだの気配を確かめていました。"},
			{event: "調子の波を見つめた", material: groupName + "で、調子の波を静かに見つめていました。"},
			{event: "動く気配に近づいた", material: groupName + "の近くで、動く気配に近づいていました。"},
		}
	case "place":
		return []reportLine{
			{event: "場所の匂いを確かめた", material: groupName + "へ向かい、場所の匂いを確かめていました。"},
			{event: "景色の気配を拾った", material: groupName + "で、景色の気配を小さく拾ってきました。"},
			{event: "知らない道をのぞいた", material: groupName + "の近くで、知らない道を少しだけのぞいていました。"},
		}
	default:
		return []reportLine{
			{event: "新しい群れへ向かった", material: groupName + "へ向かい、あたりの気配を確かめていました。"},
			{event: "知らない気配を探した", material: groupName + "で、知らない気配を少し探していました。"},
			{event: "そっと移動した", material: groupName + "へ、ペットはそっと移動していました。"},
		}
	}
}

func stayedReportLines(group domain.GroupMaster) []reportLine {
	groupName := group.DisplayName()
	switch groupCategory(group) {
	case "life":
		return []reportLine{
			{event: "暮らしの気配で過ごした", material: groupName + "で、暮らしの気配にゆっくりなじんでいました。"},
			{event: "穏やかな時間を過ごした", material: groupName + "で、穏やかな時間を静かに過ごしていました。"},
			{event: "いつもの空気に包まれた", material: groupName + "で、いつもの空気に包まれていました。"},
		}
	case "hobby":
		return []reportLine{
			{event: "好きな気配のそばにいた", material: groupName + "で、好きな気配のそばに座っていました。"},
			{event: "楽しげな空気を眺めた", material: groupName + "で、楽しげな空気をのんびり眺めていました。"},
			{event: "小さな遊びを見つけた", material: groupName + "で、小さな遊びを見つけたようです。"},
		}
	case "creation":
		return []reportLine{
			{event: "作る気配を見守った", material: groupName + "で、作る気配をそっと見守っていました。"},
			{event: "新しい形を眺めた", material: groupName + "で、新しい形が生まれる気配を眺めていました。"},
			{event: "手ざわりのある時間を過ごした", material: groupName + "で、手ざわりのある時間を過ごしていました。"},
		}
	case "work_study":
		return []reportLine{
			{event: "集中の気配で過ごした", material: groupName + "で、集中の気配に静かに寄り添っていました。"},
			{event: "学びの空気を吸った", material: groupName + "で、学びの空気を少し吸い込んでいました。"},
			{event: "机のそばで丸まった", material: groupName + "で、机のそばに丸まっていました。"},
		}
	case "digital":
		return []reportLine{
			{event: "画面の光を眺めた", material: groupName + "で、画面の光を静かに眺めていました。"},
			{event: "電子の気配で過ごした", material: groupName + "で、電子の気配に包まれていました。"},
			{event: "小さな通知の気配を聞いた", material: groupName + "で、小さな通知の気配を聞いていました。"},
		}
	case "special":
		return []reportLine{
			{event: "特別な気配で過ごした", material: groupName + "で、特別な気配をそっと抱えていました。"},
			{event: "きらめきを眺めた", material: groupName + "で、遠くのきらめきを眺めていました。"},
			{event: "いつもと違う空気にいた", material: groupName + "で、いつもと違う空気の中にいました。"},
		}
	case "thinking":
		return []reportLine{
			{event: "静かに考えごとをした", material: groupName + "で、ペットは静かに考えごとをしていました。"},
			{event: "遠くを見ていた", material: groupName + "で、遠くを見るように過ごしていました。"},
			{event: "小さな問いを抱えた", material: groupName + "で、小さな問いを抱えているようでした。"},
		}
	case "condition":
		return []reportLine{
			{event: "からだの気配で過ごした", material: groupName + "で、からだの気配に耳を澄ませていました。"},
			{event: "調子を確かめた", material: groupName + "で、自分の調子を確かめるように過ごしていました。"},
			{event: "少し身軽そうにしていた", material: groupName + "で、少し身軽そうにしていました。"},
		}
	case "place":
		return []reportLine{
			{event: "景色の中で過ごした", material: groupName + "で、景色の中に小さく座っていました。"},
			{event: "場所の空気になじんだ", material: groupName + "で、場所の空気にゆっくりなじんでいました。"},
			{event: "道の気配を眺めた", material: groupName + "で、道の気配をぼんやり眺めていました。"},
		}
	default:
		return []reportLine{
			{event: "群れで過ごした", material: groupName + "で、静かに時間を過ごしていました。"},
			{event: "気配を確かめた", material: groupName + "で、あたりの気配を確かめていました。"},
			{event: "ゆっくり滞在した", material: groupName + "で、ゆっくり滞在していました。"},
		}
	}
}

func findGroup(groups []domain.GroupMaster, id *int) *domain.GroupMaster {
	if id == nil {
		return nil
	}
	for i := range groups {
		if int(groups[i].ID()) == *id {
			return &groups[i]
		}
	}
	return nil
}

func floorToHour(t time.Time) time.Time {
	return t.Truncate(time.Hour)
}

func simulationSeed(petID string, simulatedAt time.Time) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(petID))
	_, _ = h.Write([]byte(simulatedAt.Format(time.RFC3339)))
	return int64(h.Sum64())
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
