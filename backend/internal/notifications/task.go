package notifications

import "encoding/json"

const TypeEmailLoginNotification = "email:login_notification"

type LoginNotificationPayload struct {
	Email string `json:"email"`
}

func (p *LoginNotificationPayload) Marshal() ([]byte, error) {
	return json.Marshal(p)
}
