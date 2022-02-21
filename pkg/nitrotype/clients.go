package nitrotype

type APIClient interface {
	GetTeam(tagName string) (*TeamV2APIResponse, error)
	GetProfile(username string) (*UserProfile, error)
}
