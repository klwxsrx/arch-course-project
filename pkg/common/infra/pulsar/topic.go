package pulsar

import "fmt"

func getTopicFullName(topic string) string {
	return fmt.Sprintf("persistent://public/default/%s", topic)
}
