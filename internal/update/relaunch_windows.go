package update

// Windows handles are not inherited unless marked inheritable.
func markInheritedFDsCloseOnExec() {}
