package nitrotype

import (
	"encoding/json"
	"fmt"
	"regexp"
)

var (
	NTUserProfileExtractRegExp = regexp.MustCompile(`(?m)RACER_INFO: (.*),$`)
	ErrNTUserProfileNotFound   = fmt.Errorf("unable to find nt racer data")
)

type (
	UserStatus     string
	MembershipType string
	TeamRole       string
)

const (
	UserStatusActive UserStatus = "active"
	UserStatusBanned UserStatus = "banned"

	TeamRoleMember  TeamRole = "member"
	TeamRoleOfficer TeamRole = "officer"

	MembershipTypeBasic MembershipType = "basic"
	MembershipTypeGold  MembershipType = "gold"
)

// TeamInfo contains data of a team (from TeamAPIRespones).
type TeamInfo struct {
	TeamID            int    `json:"teamID"`
	UserID            int    `json:"userID"`
	Tag               string `json:"tag"`
	TagColor          string `json:"tagColor"`
	Name              string `json:"name"`
	MinLevel          int    `json:"minLevel"`
	MinRaces          int    `json:"minRaces"`
	MinSpeed          int    `json:"minSpeed"`
	AutoRemove        int    `json:"autoRemove"`
	OtherRequirements string `json:"otherRequirements"`
	Members           int    `json:"members"`
	ActivePercent     int    `json:"activePercent"`
	Searchable        int    `json:"searchable"`
	Enrollment        string `json:"enrollment"`
	ProfileViews      int    `json:"profileViews"`
	LastActivity      int64  `json:"lastActivity"`
	LastModified      int64  `json:"lastModified"`
	CreatedStamp      int64  `json:"createdStamp"`
	Username          string `json:"username"`
	DisplayName       string `json:"displayName"`
}

// TeamStat contains data of the team's stats (from TeamAPIRespones).
type TeamStat struct {
	Board  string `json:"board"`
	Typed  int    `json:"typed"`
	Secs   int    `json:"secs"`
	Played int    `json:"played"`
	Errs   int    `json:"errs"`
	Stamp  int    `json:"stamp"`
}

// TeamMember contains data of a a team member (from TeamAPIResponse).
type TeamMember struct {
	UserID       int            `json:"userID"`
	Played       int            `json:"played"`
	Secs         int            `json:"secs"`
	Typed        int            `json:"typed"`
	Errs         int            `json:"errs"`
	JoinStamp    int64          `json:"joinStamp"`
	LastActivity int64          `json:"lastActivity"`
	Role         TeamRole       `json:"role"`
	Username     string         `json:"username"`
	DisplayName  string         `json:"displayName"`
	Membership   MembershipType `json:"membership"`
	RacesPlayed  int            `json:"racesPlayed"`
	AvgSpeed     int            `json:"avgSpeed"`
	CarID        int            `json:"carID"`
	CarHueAngle  int            `json:"carHueAngle"`
	LastLogin    int64          `json:"lastLogin"`
	Status       UserStatus     `json:"status"`
}

// TeamMemberSeason contains data of a team member's season performance.
type TeamMemberSeason struct {
	UserID       int            `json:"userID"`
	Played       int            `json:"played"`
	Secs         int            `json:"secs"`
	Typed        int            `json:"typed"`
	Errs         int            `json:"errs"`
	Points       int            `json:"points"`
	LastActivity int64          `json:"lastActivity"`
	Role         string         `json:"role"`
	Username     string         `json:"username"`
	DisplayName  string         `json:"displayName"`
	Membership   MembershipType `json:"membership"`
	RacesPlayed  int            `json:"racesPlayed"`
	AvgSpeed     int            `json:"avgSpeed"`
	Title        string         `json:"title"`
	LastLogin    int64          `json:"lastLogin"`
	Status       UserStatus     `json:"status"`
}

// TeamAPIResponse contains the response package from GET /api/teams/[team_name].
// This struct does not include extra information (eg. MOTD, applicants).
type TeamAPIResponse struct {
	Success bool `json:"success"`
	Data    struct {
		RedirectToTeam *string             `json:"redirect_to_team"`
		Info           *TeamInfo           `json:"info"`
		NoTeam         *bool               `json:"noTeam"`
		Members        []*TeamMember       `json:"members"`
		Stats          []*TeamStat         `json:"stats"`
		Season         []*TeamMemberSeason `json:"season"`
	} `json:"data"`
}

// TeamV2APIResponse contains the response package from GET /api/v2/teams/[team_name].
// This struct does not include extra information (eg. MOTD, applicants).
type TeamV2APIResponse struct {
	Status  string `json:"status"`
	Results struct {
		Info    *TeamInfo           `json:"info"`
		NoTeam  *bool               `json:"noTeam"`
		Members []*TeamMember       `json:"members"`
		Stats   []*TeamStat         `json:"stats"`
		Season  []*TeamMemberSeason `json:"season"`
	} `json:"results"`
}

// UserProfile contains data from the NT Racer profile.
type UserProfile struct {
	UserID             int            `json:"userID"`
	Username           string         `json:"username"`
	Membership         MembershipType `json:"membership"`
	DisplayName        string         `json:"displayName"`
	Title              string         `json:"title"`
	Experience         int            `json:"experience"`
	Level              int            `json:"level"`
	TeamID             *int           `json:"teamID"`
	LookingForTeam     int            `json:"lookingForTeam"`
	CarID              int            `json:"carID"`
	CarHueAngle        int            `json:"carHueAngle"`
	TotalCars          int            `json:"totalCars"`
	Nitros             int            `json:"nitros"`
	NitrosUsed         int            `json:"nitrosUsed"`
	RacesPlayed        int            `json:"racesPlaywed"`
	LongestSession     int            `json:"longestSession"`
	AvgSpeed           int            `json:"avgSpeed"`
	HighestSpeed       int            `json:"highestSpeed"`
	AllowFriendRequest int            `json:"allowFriendRequests"`
	ProfileViews       int            `json:"profileViews"`
	CreatedStamp       int            `json:"createdStamp"`
	Tag                *string        `json:"tag"`
	TagColor           *string        `json:"tagColor"`
	Garage             []string       `json:"garage"`
	Cars               []Car          `json:"cars"`
	Loot               []Loot         `json:"loot"`
}

// TODO: Enum types for loots.
// TODO: Work out car type.

// Loot contains loot information (usually used on the UserProfile).
type Loot struct {
	LootID       int        `json:"lootID"`
	Type         string     `json:"type"`
	Name         string     `json:"name"`
	AssetKey     string     `json:"assetKey"`
	Options      LootOption `json:"options"`
	Equipped     int        `json:"equipped"`
	CreatedStamp int        `json:"createdStamp"`
}

// LootOption contains options about a loot.
type LootOption struct {
	Src    string `json:"src"`
	Type   string `json:"type"`
	Rarity string `json:"rarity"`
}

// Car contains car info and it's paint job.
type Car struct {
	CarID        int
	Status       string
	CarHueAngle  int
	CreatedStamp int
}

func (c *Car) UnmarshalJSON(bs []byte) error {
	data := []interface{}{}
	err := json.Unmarshal(bs, &data)
	if err != nil {
		return err
	}
	carID, ok := data[0].(float64)
	if !ok {
		return fmt.Errorf("failed to get car id")
	}
	status, ok := data[1].(string)
	if !ok {
		return fmt.Errorf("failed to get status")
	}
	carHueAngle, ok := data[2].(float64)
	if !ok {
		return fmt.Errorf("failed to get car hue angle")
	}
	createdStamp, ok := data[3].(float64)
	if !ok {
		return fmt.Errorf("failed to get created stamp")
	}
	c.CarID = int(carID)
	c.Status = status
	c.CarHueAngle = int(carHueAngle)
	c.CreatedStamp = int(createdStamp)
	return nil
}
