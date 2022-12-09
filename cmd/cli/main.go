package main

import (
	"flag"
	"github.com/estrys/estrys/cmd"
	activitypubclient "github.com/estrys/estrys/internal/activitypub/client"
	"github.com/estrys/estrys/internal/dic"
	"github.com/estrys/estrys/internal/logger"
	"github.com/estrys/estrys/internal/repository"
	"github.com/go-fed/activity/streams"
	"net/url"
	"os"
	"time"
)

func main() {
	appContext, _, err := cmd.Bootstrap()
	log := dic.GetService[logger.Logger]()
	if err != nil {
		if log != nil {
			log.WithError(err).Error("unable to start application")
			os.Exit(1)
		}
		panic(err)
	}

	from := flag.String("from", "eliecharra", "")
	message := flag.String("message", "test message", "")

	userUrl, _ := url.Parse("https://elie.eu.ngrok.io/users/" + *from)
	user, _ := dic.GetService[repository.UserRepository]().Get(appContext, *from)

	actorUrl, _ := url.Parse("https://elie2.eu.ngrok.io/users/elie")
	actor, _ := dic.GetService[repository.ActorRepository]().
		Get(appContext, actorUrl)

	create := streams.NewActivityStreamsCreate()
	id := streams.NewJSONLDIdProperty()
	idUrl, _ := url.Parse("https://elie.eu.ngrok.io/uuid")
	id.Set(idUrl)
	create.SetJSONLDId(id)
	act := streams.NewActivityStreamsActorProperty()
	act.AppendIRI(userUrl)
	create.SetActivityStreamsActor(act)
	note := streams.NewActivityStreamsNote()
	noteAttributedTo := streams.NewActivityStreamsAttributedToProperty()
	noteAttributedTo.AppendIRI(userUrl)
	note.SetActivityStreamsAttributedTo(noteAttributedTo)
	noteContent := streams.NewActivityStreamsContentProperty()
	noteContent.AppendXMLSchemaString(*message)
	note.SetActivityStreamsContent(noteContent)
	notePublished := streams.NewActivityStreamsPublishedProperty()
	notePublished.Set(time.Now())
	note.SetActivityStreamsPublished(notePublished)
	noteTo := streams.NewActivityStreamsToProperty()
	publicUrl, _ := url.Parse("https://www.w3.org/ns/activitystreams#Public")
	noteTo.AppendIRI(publicUrl)
	note.SetActivityStreamsTo(noteTo)
	obj := streams.NewActivityStreamsObjectProperty()
	obj.AppendActivityStreamsNote(note)
	create.SetActivityStreamsObject(obj)

	//str, _ := streams.Serialize(create)
	//fmt.Printf("%+v", str)
	//os.Exit(1)

	activityPubClient := dic.GetService[activitypubclient.ActivityPubClient]()
	err = activityPubClient.PostInbox(appContext, actor, user, create)
	if err != nil {
		panic(err)
	}
}
