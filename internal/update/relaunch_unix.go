//go:build !windows

package update

import "syscall"

// markInheritedFDsCloseOnExec keeps descriptors this process inherited from
// its parent out of the relaunched app. The AppImage runtime hands its app an
// open descriptor to the image mount without close-on-exec; if the new
// version inherited it, the old mount (and its runtime process) would stay
// alive until the new version exits.
func markInheritedFDsCloseOnExec() {
	limit := uint64(1024)
	var rl syscall.Rlimit
	if syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rl) == nil && rl.Cur > limit {
		limit = min(rl.Cur, 1<<16)
	}
	for fd := 3; fd < int(limit); fd++ {
		syscall.CloseOnExec(fd)
	}
}
