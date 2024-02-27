package telemetry

import (
	"log"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/sys/windows/registry"
)

func getOrCreateUserId() string {

	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\SQMClient`, registry.ALL_ACCESS)
	if err != nil {
		log.Fatal(err)
	}
	defer k.Close()

	s, _, err := k.GetStringValue("UserId")
	if err != nil && strings.Contains(err.Error(), "The system cannot find the file specified") {
		id := uuid.New()
		k.SetStringValue("UserId", "{"+strings.ToUpper(id.String()+"}"))
		s, _, _ = k.GetStringValue("UserId")
	}
	if strings.Contains(s, "{") {
		return s[1 : len(s)-1]
	}
	return s
}
