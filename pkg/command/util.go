package command

func GetCommandReadPreferenceMode(c Command) string {
	if cr, ok := c.(CommandReadPreference); ok {
		pref := cr.GetReadPreference()
		if pref != nil {
			return pref.Mode
		}
	}
	// TODO: default option somehow?
	return "primary" // Default to primary if the command doesn't specify
}

func GetCommandDatabase(c Command) string {
	if cr, ok := c.(CommandDatabase); ok {
		return cr.GetDatabase()
	}
	return "admin"
}

func GetCommandCollection(c Command) string {
	if cr, ok := c.(CommandCollection); ok {
		return cr.GetCollection()
	}
	return ""
}
