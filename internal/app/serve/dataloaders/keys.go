package dataloaders

import (
	"fmt"
)

// Key is the interface that all keys need to implement.
type Key interface {
	// String returns a guaranteed unique string that can be used to identify an object.
	String() string
	// Raw returns the raw, underlaying value of the key.
	Raw() interface{}
}

// Keys wraps a slice of Key types to provide some convenience methods.
type Keys []Key

// Keys returns the list of strings. One for each "Key" in the list
func (l Keys) Keys() []string {
	list := make([]string, len(l))
	for i := range l {
		list[i] = l[i].String()
	}
	return list
}

////////////////////////////
//  Key: IDRaffleUserKey  //
////////////////////////////

// IDTimeRangeKey implements Key interface for an id with time range.
type IDRaffleUserKey struct {
	UserID   string
	RaffleID string
}

// String is an identity method. Used to implement String interface.
func (k IDRaffleUserKey) String() string {
	return fmt.Sprintf("%s::%s", k.UserID, k.RaffleID)
}

// Raw is an identity method. Used to implement Key Raw.
func (k IDRaffleUserKey) Row() interface{} {
	return k
}

//////////////////////
//  Key: StringKey  //
//////////////////////

// StringKey implements the Key interface for a string.
type StringKey string

// String is an identity method. Used to implement String interface.
func (k StringKey) String() string { return string(k) }

// Raw is an identity method. Used to implement Key Raw.
func (k StringKey) Raw() interface{} { return k }

// NewKeysFromStrings converts a `[]strings` to a `Keys` ([]Key).
func NewKeysFromStrings(strings []string) Keys {
	list := make(Keys, len(strings))
	for i := range strings {
		list[i] = StringKey(strings[i])
	}
	return list
}
