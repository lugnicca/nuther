//go:build !windows

package smart

// GetCommonDevicePaths returns common device paths for Unix/Linux systems
func GetCommonDevicePaths() []string {
	return []string{
		"/dev/sda", "/dev/sdb", "/dev/sdc", "/dev/sdd",
		"/dev/sde", "/dev/sdf", "/dev/sdg", "/dev/sdh",
		"/dev/nvme0n1", "/dev/nvme1n1", "/dev/nvme2n1", "/dev/nvme3n1",
		"/dev/vda", "/dev/vdb", "/dev/vdc",
		"/dev/xvda", "/dev/xvdb", "/dev/xvdc",
	}
}
