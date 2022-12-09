package tasks_test

import (
	"context"
	"encoding/pem"
	"net/http"
	"net/url"
	"os"
	"path"
	"testing"

	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/hibiken/asynq"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	activitypubclient "github.com/estrys/estrys/internal/activitypub/client"
	activitypubclientmocks "github.com/estrys/estrys/internal/activitypub/client/mocks"
	"github.com/estrys/estrys/internal/dic"
	"github.com/estrys/estrys/internal/models"
	"github.com/estrys/estrys/internal/repository"
	"github.com/estrys/estrys/internal/repository/mocks"
	"github.com/estrys/estrys/internal/worker/tasks"
	dic_test "github.com/estrys/estrys/tests/dic"
)

func TestHandleAcceptFollow(t *testing.T) {

	fakePrivateKey, _ := os.ReadFile(path.Join("testdata", "key.pem"))
	decodedPrivKeyPem, _ := pem.Decode(fakePrivateKey)

	tests := []struct {
		name      string
		inputFile string
		Mock      func(t *testing.T)
		err       string
	}{
		{
			name:      "empty payload",
			inputFile: "empty",
			err:       "unable to deserialize task input : unexpected end of JSON input: skip retry for the task",
		},
		{
			name:      "invalid activity",
			inputFile: "invalid_activity",
			err:       "unable to find a follow activity : cannot determine ActivityStreams type: 'type' property is missing: skip retry for the task",
		},
		{
			name:      "unable to fetch user",
			inputFile: "follow_ok",
			err:       "unable to fetch user from database: database is down",
			Mock: func(t *testing.T) {
				fakeUserRepo := mocks.NewUserRepository(t)
				fakeUserRepo.On("Get", mock.Anything, "fake-username").
					Return(nil, errors.New("database is down"))
				_ = dic.Register[repository.UserRepository](fakeUserRepo)
			},
		},
		{
			name:      "unable to fetch user",
			inputFile: "follow_ok",
			err:       "unable to fetch user from database: database is down",
			Mock: func(t *testing.T) {
				fakeUserRepo := mocks.NewUserRepository(t)
				fakeUserRepo.On("Get", mock.Anything, "fake-username").
					Return(nil, errors.New("database is down"))
				_ = dic.Register[repository.UserRepository](fakeUserRepo)
			},
		},
		{
			name:      "invalid actor url",
			inputFile: "follow_invalid_actor_url",
			err:       `unable parse actor URL : unable to parse actor as URL: parse "https://user:abc{DEf1=ghi@example.com": net/url: invalid userinfo: skip retry for the task`,
			Mock: func(t *testing.T) {
				fakeUserRepo := mocks.NewUserRepository(t)
				fakeUserRepo.On("Get", mock.Anything, "fake-username").
					Return(nil, nil)
				_ = dic.Register[repository.UserRepository](fakeUserRepo)
			},
		},
		{
			name:      "unable to fetch actor",
			inputFile: "follow_ok",
			err:       `unable to retrieve actor: actor does not exist`,
			Mock: func(t *testing.T) {
				fakeUserRepo := mocks.NewUserRepository(t)
				fakeUserRepo.On("Get", mock.Anything, "fake-username").
					Return(nil, nil)
				_ = dic.Register[repository.UserRepository](fakeUserRepo)

				fakeActorRepo := mocks.NewActorRepository(t)
				fakeActorUrl, _ := url.Parse("https://another-instance.example.com/users/fake-actor")
				fakeActorRepo.On("Get", mock.Anything, fakeActorUrl).
					Return(nil, errors.New("actor does not exist"))
				_ = dic.Register[repository.ActorRepository](fakeActorRepo)
			},
		},
		{
			name:      "test inbox not accepted",
			inputFile: "follow_ok",
			Mock: func(t *testing.T) {
				fakeUserRepo := mocks.NewUserRepository(t)
				fakeUser := &models.User{
					Username:   "fake-username",
					PrivateKey: decodedPrivKeyPem.Bytes,
				}
				fakeUserRepo.On("Get", mock.Anything, "fake-username").
					Return(fakeUser, nil)
				_ = dic.Register[repository.UserRepository](fakeUserRepo)

				fakeActorRepo := mocks.NewActorRepository(t)
				fakeActorUrl, _ := url.Parse("https://another-instance.example.com/users/fake-actor")
				fakeActor := &models.Actor{
					URL: "/users/fake-actor",
				}
				fakeActorRepo.On("Get", mock.Anything, fakeActorUrl).
					Return(fakeActor, nil)
				_ = dic.Register[repository.ActorRepository](fakeActorRepo)

				fakeActivityPubClient := activitypubclientmocks.NewActivityPubClient(t)
				fakeActivityPubClient.On(
					"PostInbox",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&activitypubclient.InboxNotAcceptedError{StatusCode: http.StatusUnauthorized})
				_ = dic.Register[activitypubclient.ActivityPubClient](fakeActivityPubClient)
			},
			err: "post to inbox was not accepted: skip retry for the task: error posting to inbox 401",
		},
		{
			name:      "test inbox post got a bad gateway",
			inputFile: "follow_ok",
			Mock: func(t *testing.T) {
				fakeUserRepo := mocks.NewUserRepository(t)
				fakeUser := &models.User{
					Username:   "fake-username",
					PrivateKey: decodedPrivKeyPem.Bytes,
				}
				fakeUserRepo.On("Get", mock.Anything, "fake-username").
					Return(fakeUser, nil)
				_ = dic.Register[repository.UserRepository](fakeUserRepo)

				fakeActorRepo := mocks.NewActorRepository(t)
				fakeActorUrl, _ := url.Parse("https://another-instance.example.com/users/fake-actor")
				fakeActor := &models.Actor{
					URL: "/users/fake-actor",
				}
				fakeActorRepo.On("Get", mock.Anything, fakeActorUrl).
					Return(fakeActor, nil)
				_ = dic.Register[repository.ActorRepository](fakeActorRepo)

				fakeActivityPubClient := activitypubclientmocks.NewActivityPubClient(t)
				fakeActivityPubClient.On(
					"PostInbox",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&activitypubclient.InboxNotAcceptedError{StatusCode: http.StatusBadGateway})
				_ = dic.Register[activitypubclient.ActivityPubClient](fakeActivityPubClient)
			},
			err: "post to inbox was not accepted: error posting to inbox 502",
		},
		{
			name:      "ok",
			inputFile: "follow_ok",
			Mock: func(t *testing.T) {
				fakeUserRepo := mocks.NewUserRepository(t)
				fakeUser := &models.User{
					Username:   "fake-username",
					PrivateKey: decodedPrivKeyPem.Bytes,
				}
				fakeUserRepo.On("Get", mock.Anything, "fake-username").
					Return(fakeUser, nil)
				_ = dic.Register[repository.UserRepository](fakeUserRepo)

				fakeActorRepo := mocks.NewActorRepository(t)
				fakeActorUrl, _ := url.Parse("https://another-instance.example.com/users/fake-actor")
				fakeActor := &models.Actor{
					URL: "/users/fake-actor",
				}
				fakeActorRepo.On("Get", mock.Anything, fakeActorUrl).
					Return(fakeActor, nil)
				_ = dic.Register[repository.ActorRepository](fakeActorRepo)

				fakeActivityPubClient := activitypubclientmocks.NewActivityPubClient(t)
				fakeActivityPubClient.On(
					"PostInbox",
					mock.Anything,
					fakeActor,
					fakeUser,
					mock.MatchedBy(func(act vocab.ActivityStreamsAccept) bool {
						expectedActivity := map[string]any{
							"@context": "https://www.w3.org/ns/activitystreams",
							"actor":    "https://example.com/users/fake-username",
							"id":       "https://example.com/users/fake-username",
							"object": map[string]any{
								"actor":  "https://another-instance.example.com/users/fake-actor",
								"id":     "https://another-instance.example.com/bdd01ced-d657-4847-a266-2c43e1cd8dc5",
								"object": "https://example.com/users/fake-username",
								"type":   "Follow",
							},
							"type": "Accept",
						}
						actualActivity, err := streams.Serialize(act)
						assert.NoError(t, err)
						assert.Equal(t, expectedActivity, actualActivity)
						return true
					})).Return(nil)
				_ = dic.Register[activitypubclient.ActivityPubClient](fakeActivityPubClient)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.Mock != nil {
				tt.Mock(t)
			}

			dic_test.BuildTestContainer(t)
			defer dic.ResetContainer()

			payload, err := os.ReadFile(path.Join("testdata/input", tt.inputFile+".json"))
			require.NoError(t, err)
			task := asynq.NewTask(
				tasks.TypeAcceptFollow,
				payload,
			)

			err = tasks.HandleAcceptFollow(context.TODO(), task)
			if tt.err != "" {
				require.EqualError(t, err, tt.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
