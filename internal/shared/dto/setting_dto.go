package dto

// SettingsResponse is the settings returned by the settings API. The mail password is
// write-only: only whether it is set is exposed.
type SettingsResponse struct {
	MailHost        string `json:"mail_host"`
	MailPort        int    `json:"mail_port"`
	MailUsername    string `json:"mail_username"`
	MailPasswordSet bool   `json:"mail_password_set"`
	MailFrom        string `json:"mail_from"`
	FrontendURL     string `json:"frontend_url"`
}

// UpdateSettingsInput updates only the fields that are provided.
type UpdateSettingsInput struct {
	MailHost     *string `json:"mail_host" binding:"omitempty,max=255,not_blank"`
	MailPort     *int    `json:"mail_port" binding:"omitempty,min=1,max=65535"`
	MailUsername *string `json:"mail_username" binding:"omitempty,max=255"`
	MailPassword *string `json:"mail_password" binding:"omitempty,max=255"` // "" clears the password
	MailFrom     *string `json:"mail_from" binding:"omitempty,email,max=254"`
	FrontendURL  *string `json:"frontend_url" binding:"omitempty,http_url,max=255"`
}
