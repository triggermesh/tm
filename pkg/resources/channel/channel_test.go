package channel

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/triggermesh/tm/pkg/client"
)

func TestList(t *testing.T) {
	home := os.Getenv("HOME")
	namespace := os.Getenv("NAMESPACE")
	channelClient, err := client.NewClient(home + "/.tm/config.json")
	assert.NoError(t, err)
	log.Println(channelClient)

	channel := &Channel{Namespace: namespace}

	channelList, err := channel.List(&channelClient)
	assert.NoError(t, err)
	log.Println(channelList)
}

func TestBuild(t *testing.T) {
	home := os.Getenv("HOME")
	namespace := os.Getenv("NAMESPACE")
	channelClient, err := client.NewClient(home + "/.tm/config.json")
	assert.NoError(t, err)

	testCases := []struct {
		Name        string
		Provisioner string
		ErrMSG      error
	}{
		{"foo", "bar", nil},
	}

	for _, tt := range testCases {
		channel := &Channel{
			Name:        tt.Name,
			Namespace:   namespace,
			Provisioner: tt.Provisioner,
		}

		err = channel.Deploy(&channelClient)
		if err != nil {
			assert.Error(t, err)
			continue
		}

		ch, err := channel.Get(&channelClient)
		assert.NoError(t, err)
		assert.Equal(t, tt.Name, ch.Name)

		err = channel.Deploy(&channelClient)
		if err != nil {
			assert.Error(t, err)
		}

		err = channel.Delete(&channelClient)
		assert.NoError(t, err)
	}
}
