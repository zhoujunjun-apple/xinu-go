package include

// IsBadPid function checks if pid is valid or not.
// True: pid is invalid;
// False: pid is valid;
func IsBadPid(pid Pid32) bool {
	return pid < 0 || int(pid) >= NPROC // TODO: add process status check
}
