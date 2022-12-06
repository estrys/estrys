//nolint:ireturn
package activitypub

import (
	"context"
	"net/url"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/pkg/errors"

	"github.com/estrys/estrys/internal/domain/domainmodels"
	"github.com/estrys/estrys/internal/models"
	"github.com/estrys/estrys/internal/router/routes"
	"github.com/estrys/estrys/internal/router/urlgenerator"
)

type VocabService interface {
	GetActor(*domainmodels.User) (map[string]any, error)
	GetFollowers(*domainmodels.User) (map[string]any, error)
	GetFollowing(*domainmodels.User) (map[string]any, error)
	GetOutbox(*domainmodels.User) (map[string]any, error)
	GetAccept(
		user *models.User,
		act streams.ActivityStreamsInterface,
	) (vocab.ActivityStreamsAccept, error)
	GetReject(
		user *models.User,
		act streams.ActivityStreamsInterface,
	) (vocab.ActivityStreamsReject, error)
}

type activityPubService struct {
	URLGenerator urlgenerator.URLGenerator
}

func NewActivityPubVocabService(
	urlGenerator urlgenerator.URLGenerator,
) *activityPubService {
	return &activityPubService{
		URLGenerator: urlGenerator,
	}
}

func (a *activityPubService) serialize(vocab vocab.Type) (map[string]any, error) {
	serializedStreams, err := streams.Serialize(vocab)
	if err != nil {
		return nil, errors.Wrap(err, "unable to serialize vocab")
	}
	return serializedStreams, nil
}

func (a *activityPubService) GetFollowers(user *domainmodels.User) (map[string]any, error) {
	followersURL, err := a.URLGenerator.URL(
		routes.UserFollowersRoute,
		[]string{"username", user.Username},
		urlgenerator.OptionAbsoluteURL,
	)
	if err != nil {
		return nil, errors.Wrap(err, "cannot generate followers URL")
	}

	collection := streams.NewActivityStreamsOrderedCollection()
	totalItems := streams.NewActivityStreamsTotalItemsProperty()
	totalItems.Set(int(user.Metrics.Followers))
	id := streams.NewJSONLDIdProperty()
	id.Set(followersURL)
	collection.SetJSONLDId(id)
	collection.SetActivityStreamsTotalItems(totalItems)

	return a.serialize(collection)
}

func (a *activityPubService) GetFollowing(user *domainmodels.User) (map[string]any, error) {
	followingURL, err := a.URLGenerator.URL(
		routes.UserFollowingRoute,
		[]string{"username", user.Username},
		urlgenerator.OptionAbsoluteURL,
	)
	if err != nil {
		return nil, errors.Wrap(err, "cannot generate following URL")
	}

	collection := streams.NewActivityStreamsOrderedCollection()
	totalItems := streams.NewActivityStreamsTotalItemsProperty()
	totalItems.Set(int(user.Metrics.Following))
	id := streams.NewJSONLDIdProperty()
	id.Set(followingURL)
	collection.SetJSONLDId(id)
	collection.SetActivityStreamsTotalItems(totalItems)

	return a.serialize(collection)
}

func (a *activityPubService) GetOutbox(user *domainmodels.User) (map[string]any, error) {
	outboxURL, err := a.URLGenerator.URL(
		routes.UserOutbox,
		[]string{"username", user.Username},
		urlgenerator.OptionAbsoluteURL,
	)
	if err != nil {
		return nil, errors.Wrap(err, "cannot generate outbox URL")
	}

	collection := streams.NewActivityStreamsOrderedCollection()
	totalItems := streams.NewActivityStreamsTotalItemsProperty()
	totalItems.Set(int(user.Metrics.Tweets))
	id := streams.NewJSONLDIdProperty()
	id.Set(outboxURL)
	collection.SetJSONLDId(id)
	collection.SetActivityStreamsTotalItems(totalItems)

	return a.serialize(collection)
}

func (a *activityPubService) GetActor(user *domainmodels.User) (map[string]any, error) {
	actor := streams.NewActivityStreamsService()

	outboxURL, err := a.URLGenerator.URL(
		routes.UserOutbox,
		[]string{"username", user.Username},
		urlgenerator.OptionAbsoluteURL,
	)
	if err != nil {
		return nil, errors.Wrap(err, "cannot generate outbox URL")
	}
	outbox := streams.NewActivityStreamsOutboxProperty()
	outbox.SetIRI(outboxURL)
	actor.SetActivityStreamsOutbox(outbox)

	followersURL, err := a.URLGenerator.URL(
		routes.UserFollowersRoute,
		[]string{"username", user.Username},
		urlgenerator.OptionAbsoluteURL,
	)
	if err != nil {
		return nil, errors.Wrap(err, "cannot generate outbox URL")
	}
	followers := streams.NewActivityStreamsFollowersProperty()
	followers.SetIRI(followersURL)
	actor.SetActivityStreamsFollowers(followers)

	followingURL, err := a.URLGenerator.URL(
		routes.UserFollowingRoute,
		[]string{"username", user.Username},
		urlgenerator.OptionAbsoluteURL,
	)
	if err != nil {
		return nil, errors.Wrap(err, "cannot generate outbox URL")
	}
	following := streams.NewActivityStreamsFollowingProperty()
	following.SetIRI(followingURL)
	actor.SetActivityStreamsFollowing(following)

	inboxURL, _ := url.Parse("https://elie.eu.ngrok.io/users/smaftoul/inbox")
	inbox := streams.NewActivityStreamsInboxProperty()
	inbox.SetIRI(inboxURL)
	actor.SetActivityStreamsInbox(inbox)

	username := streams.NewActivityStreamsPreferredUsernameProperty()
	username.SetXMLSchemaString(user.Username)
	actor.SetActivityStreamsPreferredUsername(username)

	name := streams.NewActivityStreamsNameProperty()
	name.AppendXMLSchemaString(user.Name)
	actor.SetActivityStreamsName(name)

	summary := streams.NewActivityStreamsSummaryProperty()
	summary.AppendXMLSchemaString(user.Description)
	actor.SetActivityStreamsSummary(summary)

	published := streams.NewActivityStreamsPublishedProperty()
	published.Set(user.CreatedAt)
	actor.SetActivityStreamsPublished(published)

	userURL, err := a.URLGenerator.URL(
		routes.UserRoute,
		[]string{"username", user.Username},
		urlgenerator.OptionAbsoluteURL,
	)
	if err != nil {
		return nil, errors.Wrap(err, "cannot generate outbox URL")
	}
	userID := streams.NewJSONLDIdProperty()
	userID.Set(userURL)

	publicKeyProp := streams.NewW3IDSecurityV1PublicKeyProperty()
	pubKey := streams.NewW3IDSecurityV1PublicKey()
	publicKeyPem := streams.NewW3IDSecurityV1PublicKeyPemProperty()
	publicKeyPem.Set(user.PublicKeyPem())
	pubKeyowner := streams.NewW3IDSecurityV1OwnerProperty()
	pubKeyowner.Set(userURL)
	pubKey.SetJSONLDId(userID)
	pubKey.SetW3IDSecurityV1Owner(pubKeyowner)
	pubKey.SetW3IDSecurityV1PublicKeyPem(publicKeyPem)
	publicKeyProp.AppendW3IDSecurityV1PublicKey(pubKey)
	actor.SetW3IDSecurityV1PublicKey(publicKeyProp)

	icon := streams.NewActivityStreamsIconProperty()
	image := streams.NewActivityStreamsImage()
	profileImageURL := streams.NewActivityStreamsUrlProperty()
	profileImageURL.AppendIRI(user.ProfileImageURL)
	image.SetActivityStreamsUrl(profileImageURL)
	icon.AppendActivityStreamsImage(image)
	actor.SetActivityStreamsIcon(icon)

	actor.SetJSONLDId(userID)

	return a.serialize(actor)
}

func (a *activityPubService) GetAccept(
	user *models.User,
	act streams.ActivityStreamsInterface,
) (vocab.ActivityStreamsAccept, error) {
	userURL, err := a.URLGenerator.URL(
		routes.UserRoute,
		[]string{"username", user.Username},
		urlgenerator.OptionAbsoluteURL,
	)
	if err != nil {
		return nil, errors.Wrap(err, "cannot generate user URL")
	}
	userID := streams.NewJSONLDIdProperty()
	userID.Set(userURL)

	acceptActivity := streams.NewActivityStreamsAccept()
	acceptActivity.SetJSONLDId(userID)
	activityPubActor := streams.NewActivityStreamsActorProperty()
	activityPubActor.AppendIRI(userURL)
	acceptActivity.SetActivityStreamsActor(activityPubActor)

	object := streams.NewActivityStreamsObjectProperty()
	resolver, err := streams.NewTypeResolver(
		func(ctx context.Context, follow vocab.ActivityStreamsFollow) error {
			object.AppendActivityStreamsFollow(follow)
			return nil
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create type resolver")
	}
	err = resolver.Resolve(context.Background(), act)
	if err != nil {
		return nil, errors.Wrap(err, "unable to resolve activity")
	}

	if object.Len() == 0 {
		return nil, errors.Errorf("unsupported activity %s", act.GetTypeName())
	}
	acceptActivity.SetActivityStreamsObject(object)

	return acceptActivity, nil
}

func (a *activityPubService) GetReject(
	user *models.User,
	act streams.ActivityStreamsInterface,
) (vocab.ActivityStreamsReject, error) {
	userURL, err := a.URLGenerator.URL(
		routes.UserRoute,
		[]string{"username", user.Username},
		urlgenerator.OptionAbsoluteURL,
	)
	if err != nil {
		return nil, errors.Wrap(err, "cannot generate user URL")
	}
	userID := streams.NewJSONLDIdProperty()
	userID.Set(userURL)

	acceptActivity := streams.NewActivityStreamsReject()
	acceptActivity.SetJSONLDId(userID)
	activityPubActor := streams.NewActivityStreamsActorProperty()
	activityPubActor.AppendIRI(userURL)
	acceptActivity.SetActivityStreamsActor(activityPubActor)

	object := streams.NewActivityStreamsObjectProperty()
	resolver, err := streams.NewTypeResolver(
		func(ctx context.Context, follow vocab.ActivityStreamsFollow) error {
			object.AppendActivityStreamsFollow(follow)
			return nil
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create type resolver")
	}
	err = resolver.Resolve(context.Background(), act)
	if err != nil {
		return nil, errors.Wrap(err, "unable to resolve activity")
	}

	if object.Len() == 0 {
		return nil, errors.Errorf("unsupported activity %s", act.GetTypeName())
	}
	acceptActivity.SetActivityStreamsObject(object)

	return acceptActivity, nil
}

type withObject interface {
	GetActivityStreamsObject() vocab.ActivityStreamsObjectProperty
}

func GetActorURL(act pub.Activity) (*url.URL, error) {
	actorStr, err := act.GetActivityStreamsActor().Serialize()
	if err != nil {
		return nil, errors.Wrap(err, "unable read activity streams actor field")
	}
	actorURL, err := url.Parse(actorStr.(string))
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse actor as URL")
	}
	return actorURL, nil
}

func GetObjectURL(act withObject) (*url.URL, error) {
	objectStr, err := act.GetActivityStreamsObject().Serialize()
	if err != nil {
		return nil, errors.Wrap(err, "unable read activity streams object field")
	}
	objectURL, err := url.Parse(objectStr.(string))
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse follow object as URL")
	}
	return objectURL, nil
}
