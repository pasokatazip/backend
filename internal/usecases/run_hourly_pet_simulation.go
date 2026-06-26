package usecases

import (
	"hash/fnv"
	"math"
	"math/rand"
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

	output := RunHourlyPetSimulationOutput{
		SimulatedAt: simulatedAt,
		TotalPets:   len(pets),
		Results:     make([]RunHourlyPetSimulationPetResult, 0, len(pets)),
	}

	for _, pet := range pets {
		// 行動計画の作成
		plan := u.planPetHour(pet, groups, simulatedAt)
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
func (u *RunHourlyPetSimulation) planPetHour(pet domain.SimulationPet, groups []domain.GroupMaster, simulatedAt time.Time) petHourPlan {
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
		nextGroup = chooseNextGroup(pet, groups, currentGroup.ID(), metrics.restNeed, r)
		moved = currentGroupID == nil || nextGroup.ID() != currentGroup.ID()
	}

	interactionCount := calculateInteractionCount(pet, nextGroup, r)
	ambientEvent, reportMaterial := buildAmbientText(nextGroup, moved, metrics.restNeed, interactionCount)
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
			PetID:            pet.ID(),
			PreviousGroupID:  previousGroupID,
			NextGroupID:      nextGroup.ID(),
			PreviousJoinID:   pet.CurrentJoinID,
			MoveReason:       "hourly_simulation",
			Moved:            moved,
			EnergyDelta:      nextGroup.EnergyDelta(),
			CuriosityDelta:   nextGroup.CuriosityDelta(),
			SocialityDelta:   nextGroup.SocialityDelta(),
			RoutineDelta:     nextGroup.RoutineDelta(),
			ExperienceAmount: 1 + interactionCount,
			SimulatedAt:      simulatedAt,
			Log:              log,
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
	restNeed := clamp(float64(45-pet.Energy())/45, 0, 1)
	currentGroupFit := calculateGroupFit(pet.Pet, group)
	attachment := clamp(float64(pet.Sociality())/500+hoursInGroup/120, 0, 0.25)
	recentMovePenalty := 0.0
	if pet.JoinedAt != nil && simulatedAt.Sub(*pet.JoinedAt) < 3*time.Hour {
		recentMovePenalty = 0.18
	}
	curiosityBonus := clamp((float64(pet.Curiosity())-50)/250, -0.08, 0.20)

	moveProbability := clamp(
		0.10+boredom+curiosityBonus+(1-currentGroupFit)*0.20+restNeed*0.25-attachment-recentMovePenalty,
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

func needWeight(status int, delta float64) float64 {
	return clamp(float64(status)/100, 0, 1) * delta * 1.5
}

// 移動先の群れを選択
func chooseNextGroup(pet domain.SimulationPet, groups []domain.GroupMaster, currentID domain.GroupMasterID, restNeed float64, r *rand.Rand) domain.GroupMaster {
	best := groups[0]
	bestScore := math.Inf(-1)
	for _, group := range groups {
		if group.ID() == currentID && len(groups) > 1 {
			continue
		}

		score := calculateGroupFit(pet.Pet, group) + r.Float64()*0.18
		if restNeed > 0.55 {
			score += math.Max(0, group.EnergyDelta()) * 2
			score += math.Max(0, group.RoutineDelta()) * 1.2
		}
		if score > bestScore {
			best = group
			bestScore = score
		}
	}
	return best
}

// 交流会数を決める
func calculateInteractionCount(pet domain.SimulationPet, group domain.GroupMaster, r *rand.Rand) int {
	base := clamp(float64(pet.Sociality())/100, 0, 1)
	if group.SocialityDelta() > 0 {
		base += group.SocialityDelta()
	}
	if r.Float64() < base*0.45 {
		return 1
	}
	return 0
}

// レポート用の文章を作成
func buildAmbientText(group domain.GroupMaster, moved bool, restNeed float64, interactionCount int) (string, string) {
	if restNeed > 0.65 {
		event := "少し休みたそうにしていた"
		return event, group.DisplayName() + "で、ペットは少し休みたそうに丸まっていました。"
	}
	if interactionCount > 0 {
		event := "近くの気配と遊んだ"
		return event, group.DisplayName() + "で、近くのYoYoと少しだけ遊んでいました。"
	}
	if moved {
		event := "新しい群れへ向かった"
		return event, group.DisplayName() + "へ向かい、あたりの気配を確かめていました。"
	}
	event := "群れで過ごした"
	return event, group.DisplayName() + "で、静かに時間を過ごしていました。"
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
