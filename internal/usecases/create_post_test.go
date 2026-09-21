package usecases

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

const (
	testCreatePostUserID     = domain.UserID("4c0c926e-6a13-4bf4-8ae4-593c4047280f")
	testCreatePostOtherID    = domain.UserID("9cb459e5-a503-4847-8988-fcceaa64277b")
	testCreatePostPetID      = domain.PetID("265faa23-d586-488a-b159-875d8be74f1f")
	testCreatePostOtherPetID = domain.PetID("1add695f-d572-47f0-a0ab-89099fdd351d")
)

type createPostRepositoryStub struct {
	called bool
	post   domain.Post
	err    error
}

func (r *createPostRepositoryStub) Create(_ context.Context, post domain.Post) (domain.Post, error) {
	r.called = true
	r.post = post
	if r.err != nil {
		return domain.Post{}, r.err
	}
	return post, nil
}

func (r *createPostRepositoryStub) FindByPetID(_ context.Context, _ domain.PetID) ([]domain.Post, error) {
	return nil, nil
}

type createPostPetRepositoryStub struct {
	workflow  *feedWorkflowStub
	pet       domain.Pet
	activePet domain.Pet
	findErr   error
	activeErr error
}

func (r *createPostPetRepositoryStub) Create(_ context.Context, _ domain.Pet) (domain.Pet, error) {
	return domain.Pet{}, nil
}
func (r *createPostPetRepositoryStub) FindByID(_ context.Context, _ domain.PetID) (domain.Pet, error) {
	return r.pet, r.findErr
}
func (r *createPostPetRepositoryStub) FindActiveByUserID(_ context.Context, _ domain.UserID) (domain.Pet, error) {
	return r.activePet, r.activeErr
}
func (r *createPostPetRepositoryStub) FindAllByUserID(_ context.Context, _ domain.UserID) ([]domain.Pet, error) {
	return nil, nil
}
func (r *createPostPetRepositoryStub) FindDeletedByUserID(_ context.Context, _ domain.UserID) ([]domain.Pet, error) {
	return nil, nil
}
func (r *createPostPetRepositoryStub) UpdateProfile(_ context.Context, _ domain.PetID, _ domain.UserID, _ string, _ string, _ time.Time) (domain.Pet, error) {
	return domain.Pet{}, nil
}

func newCreatePostTestPet(id domain.PetID, userID domain.UserID, deleted bool) domain.Pet {
	return domain.NewPet(id, "pet", domain.DefaultPetColor, deleted, userID, 0, 0, 0, 0, nil, 1, time.Time{}, time.Time{})
}

func TestCreatePostExecute(t *testing.T) {
	ownedPet := newCreatePostTestPet(testCreatePostPetID, testCreatePostUserID, false)

	tests := []struct {
		name        string
		input       CreatePostInput
		petRepo     *createPostPetRepositoryStub
		wantErr     error
		wantCreate  bool
		wantContent string
	}{
		{
			name:        "creates a post for the authenticated user's active pet",
			input:       CreatePostInput{UserID: testCreatePostUserID, PetID: testCreatePostPetID, Content: "hello"},
			petRepo:     &createPostPetRepositoryStub{pet: ownedPet, activePet: ownedPet},
			wantCreate:  true,
			wantContent: "hello",
		},
		{
			name:        "accepts one character and trims surrounding whitespace",
			input:       CreatePostInput{UserID: testCreatePostUserID, PetID: testCreatePostPetID, Content: "  あ  "},
			petRepo:     &createPostPetRepositoryStub{pet: ownedPet, activePet: ownedPet},
			wantCreate:  true,
			wantContent: "あ",
		},
		{
			name:    "rejects empty content",
			input:   CreatePostInput{UserID: testCreatePostUserID, PetID: testCreatePostPetID, Content: ""},
			petRepo: &createPostPetRepositoryStub{},
			wantErr: domain.ErrValidation,
		},
		{
			name:    "rejects whitespace-only content",
			input:   CreatePostInput{UserID: testCreatePostUserID, PetID: testCreatePostPetID, Content: " \n\t "},
			petRepo: &createPostPetRepositoryStub{},
			wantErr: domain.ErrValidation,
		},
		{
			name:    "rejects an invalid user ID",
			input:   CreatePostInput{UserID: "invalid", PetID: testCreatePostPetID, Content: "hello"},
			petRepo: &createPostPetRepositoryStub{},
			wantErr: domain.ErrValidation,
		},
		{
			name:    "rejects an invalid pet ID",
			input:   CreatePostInput{UserID: testCreatePostUserID, PetID: "invalid", Content: "hello"},
			petRepo: &createPostPetRepositoryStub{},
			wantErr: domain.ErrValidation,
		},
		{
			name:    "returns not found when the pet does not exist",
			input:   CreatePostInput{UserID: testCreatePostUserID, PetID: testCreatePostPetID, Content: "hello"},
			petRepo: &createPostPetRepositoryStub{findErr: domain.ErrNotFound},
			wantErr: domain.ErrNotFound,
		},
		{
			name:    "rejects a pet owned by another user",
			input:   CreatePostInput{UserID: testCreatePostUserID, PetID: testCreatePostPetID, Content: "hello"},
			petRepo: &createPostPetRepositoryStub{pet: newCreatePostTestPet(testCreatePostPetID, testCreatePostOtherID, false)},
			wantErr: domain.ErrUnauthorized,
		},
		{
			name:    "rejects a deleted pet",
			input:   CreatePostInput{UserID: testCreatePostUserID, PetID: testCreatePostPetID, Content: "hello"},
			petRepo: &createPostPetRepositoryStub{pet: newCreatePostTestPet(testCreatePostPetID, testCreatePostUserID, true)},
			wantErr: domain.ErrValidation,
		},
		{
			name:  "rejects an inactive pet",
			input: CreatePostInput{UserID: testCreatePostUserID, PetID: testCreatePostPetID, Content: "hello"},
			petRepo: &createPostPetRepositoryStub{
				pet:       ownedPet,
				activePet: newCreatePostTestPet(testCreatePostOtherPetID, testCreatePostUserID, false),
			},
			wantErr: domain.ErrValidation,
		},
		{
			name:  "rejects posting when the user has no active pet",
			input: CreatePostInput{UserID: testCreatePostUserID, PetID: testCreatePostPetID, Content: "hello"},
			petRepo: &createPostPetRepositoryStub{
				pet:       ownedPet,
				activeErr: domain.ErrNotFound,
			},
			wantErr: domain.ErrValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postRepo := &createPostRepositoryStub{}
			usecase := NewCreatePost(&createPostTransactionStub{}, tt.petRepo, postRepo, &createPostExperienceStub{}, &createPostEventStub{}, &createPostRuleStub{}, &feedEvolutionStub{})

			post, err := usecase.Execute(context.Background(), tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Execute() error = %v, want %v", err, tt.wantErr)
			}
			if postRepo.called != tt.wantCreate {
				t.Fatalf("Create() called = %v, want %v", postRepo.called, tt.wantCreate)
			}
			if tt.wantCreate && (post.PetID() != tt.input.PetID || post.Content() != tt.wantContent) {
				t.Fatalf("Execute() post = (%s, %q), want (%s, %q)", post.PetID(), post.Content(), tt.input.PetID, tt.wantContent)
			}
		})
	}
}

type createPostTransactionStub struct {
	err       error
	committed bool
}

func (s *createPostTransactionStub) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	if err := fn(context.WithValue(ctx, feedTransactionKey{}, true)); err != nil {
		return err
	}
	if s.err != nil {
		return s.err
	}
	s.committed = true
	return nil
}
func (r *createPostExperienceStub) AddFeedExperience(ctx context.Context, petID domain.PetID, amount int, at time.Time) (int, int, error) {
	if r.workflow != nil {
		return r.workflow.AddFeedExperience(ctx, petID, amount, at)
	}
	if amount != feedExperienceAmount {
		return 0, 0, errors.New("unexpected experience amount")
	}
	return 7, 3, nil
}
func (r *createPostRuleStub) FindSatisfiedAfterFeed(ctx context.Context, petID domain.PetID, at time.Time) (*domain.SatisfiedEvolutionRule, error) {
	if r.workflow != nil {
		return r.workflow.FindSatisfiedAfterFeed(ctx, petID, at)
	}
	return nil, nil
}
func (r *createPostPetRepositoryStub) UpdateEvolutionStage(ctx context.Context, petID domain.PetID, stageID domain.EvolutionStageID, at time.Time) error {
	if r.workflow != nil {
		return r.workflow.UpdateEvolutionStage(ctx, petID, stageID, at)
	}
	return nil
}

type createPostEventStub struct {
	domain.PetExperienceEventRepository
	event domain.PetExperienceEvent
	err   error
}

func (r *createPostEventStub) CreateInTransaction(_ context.Context, event domain.PetExperienceEvent) error {
	r.event = event
	return r.err
}

type feedWorkflowStub struct {
	*createPostRepositoryStub
	calls   []string
	failAt  string
	failure error
	rule    *domain.SatisfiedEvolutionRule
	applied domain.EvolutionStageID
}

type feedTransactionKey struct{}

func (s *feedWorkflowStub) step(ctx context.Context, name string) error {
	if ctx.Value(feedTransactionKey{}) != true {
		return errors.New("missing transaction context")
	}
	s.calls = append(s.calls, name)
	if s.failAt == name {
		return s.failure
	}
	return nil
}
func (s *feedWorkflowStub) Create(ctx context.Context, post domain.Post) (domain.Post, error) {
	s.createPostRepositoryStub.Create(ctx, post)
	return post, s.step(ctx, "post")
}
func (s *feedWorkflowStub) AddFeedExperience(ctx context.Context, petID domain.PetID, amount int, at time.Time) (int, int, error) {
	if petID != s.post.PetID() || amount != 10 || !at.Equal(s.post.CreatedAt()) {
		return 0, 0, errors.New("incorrect feed arguments")
	}
	return 7, 3, s.step(ctx, "experience")
}
func (s *feedWorkflowStub) FindSatisfiedAfterFeed(ctx context.Context, petID domain.PetID, at time.Time) (*domain.SatisfiedEvolutionRule, error) {
	if petID != s.post.PetID() || !at.Equal(s.post.CreatedAt()) {
		return nil, errors.New("incorrect rule arguments")
	}
	return s.rule, s.step(ctx, "rule")
}
func (s *feedWorkflowStub) UpdateEvolutionStage(ctx context.Context, petID domain.PetID, stage domain.EvolutionStageID, at time.Time) error {
	if petID != s.post.PetID() || !at.Equal(s.post.CreatedAt()) {
		return errors.New("incorrect evolution arguments")
	}
	s.applied = stage
	return s.step(ctx, "stage")
}

type feedEventStub struct {
	domain.PetExperienceEventRepository
	workflow *feedWorkflowStub
	event    domain.PetExperienceEvent
}

func (s *feedEventStub) CreateInTransaction(ctx context.Context, event domain.PetExperienceEvent) error {
	s.event = event
	return s.workflow.step(ctx, "event")
}
func TestCreatePostFeedWorkflow(t *testing.T) {
	failure := errors.New("write failed")
	for _, tc := range []struct {
		name, failAt string
		evolve       bool
		wantCalls    []string
	}{
		{"without evolution", "", false, []string{"post", "experience", "event", "rule"}},
		{"with evolution", "", true, []string{"post", "experience", "event", "rule", "stage", "evolution"}},
		{"post failure", "post", true, []string{"post"}},
		{"experience failure", "experience", true, []string{"post", "experience"}},
		{"event failure", "event", true, []string{"post", "experience", "event"}},
		{"rule failure", "rule", true, []string{"post", "experience", "event", "rule"}},
		{"stage failure", "stage", true, []string{"post", "experience", "event", "rule", "stage"}},
		{"evolution failure", "evolution", true, []string{"post", "experience", "event", "rule", "stage", "evolution"}},
		{"commit failure", "commit", true, []string{"post", "experience", "event", "rule", "stage", "evolution"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			workflow := &feedWorkflowStub{createPostRepositoryStub: &createPostRepositoryStub{}, failAt: tc.failAt, failure: failure}
			if tc.evolve {
				workflow.rule = &domain.SatisfiedEvolutionRule{RuleID: 2, ToStageID: 3, PrimaryStatus: "curiosity"}
			}
			events := &feedEventStub{workflow: workflow}
			transaction := &createPostTransactionStub{}
			evolutions := &feedEvolutionStub{workflow: workflow}
			if tc.failAt == "commit" {
				transaction.err = failure
			}
			pet := newCreatePostTestPet(testCreatePostPetID, testCreatePostUserID, false)
			uc := NewCreatePost(transaction, &createPostPetRepositoryStub{pet: pet, activePet: pet, workflow: workflow}, workflow, &createPostExperienceStub{workflow: workflow}, events, &createPostRuleStub{workflow: workflow}, evolutions)
			post, err := uc.Execute(context.Background(), CreatePostInput{UserID: testCreatePostUserID, PetID: testCreatePostPetID, Content: "hello"})
			if tc.failAt != "" {
				if !errors.Is(err, failure) || post.ID() != "" || transaction.committed {
					t.Fatalf("failure returned post=%v err=%v committed=%v", post, err, transaction.committed)
				}
			} else {
				if err != nil || !transaction.committed {
					t.Fatalf("success err=%v committed=%v", err, transaction.committed)
				}
				event := events.event
				if event.ID() == "" || event.PetID() != post.PetID() || event.SourceType() != domain.ExperienceSourceTypeFeed || event.SourceID() == nil || *event.SourceID() != string(post.ID()) || event.Amount() != 10 || event.CappedAmount() != 3 || !event.CreatedAt().Equal(post.CreatedAt()) || !event.ExperienceDate().Equal(post.CreatedAt()) {
					t.Fatalf("incorrect event: %+v", event)
				}
				if tc.evolve {
					evolution := evolutions.evolution
					if workflow.applied != workflow.rule.ToStageID || evolution.ID() == "" || evolution.PetID() != post.PetID() || evolution.StageID() != workflow.rule.ToStageID || evolution.EvolutionRuleID() == nil || *evolution.EvolutionRuleID() != workflow.rule.RuleID || evolution.PrimaryStatus() == nil || *evolution.PrimaryStatus() != workflow.rule.PrimaryStatus || !evolution.EvolvedAt().Equal(post.CreatedAt()) || !evolution.CreatedAt().Equal(post.CreatedAt()) {
						t.Fatalf("incorrect evolution: %+v", evolution)
					}
				} else if evolutions.evolution.ID() != "" {
					t.Fatal("unexpected evolution")
				}

			}
			if !reflect.DeepEqual(workflow.calls, tc.wantCalls) {
				t.Fatalf("calls=%v want=%v", workflow.calls, tc.wantCalls)
			}
		})
	}
}

type feedEvolutionStub struct {
	domain.PetEvolutionRepository
	workflow  *feedWorkflowStub
	evolution domain.PetEvolution
}

func (s *feedEvolutionStub) CreateInTransaction(ctx context.Context, evolution domain.PetEvolution) error {
	s.evolution = evolution
	return s.workflow.step(ctx, "evolution")
}

type createPostExperienceStub struct {
	domain.PetExperienceRepository
	workflow *feedWorkflowStub
}

type createPostRuleStub struct {
	domain.EvolutionRuleRepository
	workflow *feedWorkflowStub
}
