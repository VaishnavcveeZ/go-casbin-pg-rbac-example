package main

const (
	admin   = iota + 1 // Zero = 1
	teacher            // One = 2
	student            // Two = 3
)

var allRoles = map[int]bool{
	admin:   true,
	teacher: true,
	student: true,
}

func ValidateRole(r int) bool {
	v, ok := allRoles[r]
	return v && ok
}
