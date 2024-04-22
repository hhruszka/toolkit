package main

import "strings"

// Logic:
// - each entry can be either a test case or an enumeration.
// - each entry can have number of prerequisites that are executed recurrently
// - prerequisite can only state false or true or provide stdout
// - only test cases marked as "not test" case can be on a list of dependencies
// - test case decides about its result, prerequisite delivers only information (data, or execution result)
// - if a test case cannot be executed due to missing binary on the SUT then such a test case is marked as passed.
// In the future the latter behaviour might be changed/modified by introducing configuration parameters that would decide
// about a result (failed, passed) in such cases.

//	var EnumCases = []*TestCase{
//		{"ENUM01", false, "PATH variables defined inside /etc", []*Dependency{nil}, []string{sh, grep, tr}, `for p in $(grep -ERh "^ *PATH=.*" /etc/ 2> /dev/null | tr -d '\\x22\\x27' | cut -d= -f2 | tr ":" "\n" | sort -u); do [ -d "$p" ] && echo "$p";done`},
//		{"ENUM02", false, "Writable files outside user's home (non-root users)", []*Dependency{nil}, []string{sh, grep, tr}, `find_opt='-path /proc -prune -o -path /sys -prune -o -path /dev -prune -o';find / -path "$HOME" -prune -o $find_opts -not -type l -writable -print;find  / -path "$HOME" -prune -o $find_opts -type l -user $USER -print`},
//	}
const (
	cut    = "cut"
	find   = "find"
	getcap = "getcap"
	grep   = "grep"
	groups = "groups"
	id     = "id"
	ls     = "ls"
	sh     = "sh"
	sort   = "sort"
	sudo   = "sudo"
	tr     = "tr"
)

type DependencyType int

const (
	// ExitCode dependency is in the situation when the execution is important e.g. something exists or
	// the condition test matters
	ExitCode DependencyType = iota
	// Stdout dependency is about stdio data only since there are cases when exit code is 1 because of
	// permissions denials. 'find' command is such an example. It returns data even when the exit code is 1.
	Stdout
	TestFunc
)

// Dependency struct keeps information about a test case dependency. The test case dependency might be
// on stdout data or exit code. Stdout is an output from the command stored in a shell variable that
// the test case uses to execute test e.g. 'find' output. ExitCode is a dependency on a successful execution
// of a command and its value is stored in a shell variable as true or false.
type Dependency struct {
	Id      string
	Type    DependencyType
	VarName string
}

// TestCase type struct defines a test case.
type TestCase struct {
	// Id defines a unique test id used also as a key to cache map.
	Id string `json:"Id"`
	// IsTest is a boolean flag to denote that an entry is a test case (true). An entry can be also an enumeration (false).
	IsTest bool `json:"IsTest"`
	// Abstract is a test short description
	Abstract string `json:"Abstract"`
	// Dependencies is a slice with enumeration entries that need to be executed prior a test case
	Dependencies []*Dependency `json:"Dependencies"`
	// UsedBins is a slice with used binaries by a test case or an enumeration. It is used to verify whether
	// en entry can be executed on a container.
	UsedBins []string `json:"UsedBins"`
	// Command is a shell script implementing a test case or an enumeration that will be executed in the container.
	NonRoot    bool   // if true need to be run by non-root user
	Command    string `json:"Command"`
	stdOutVar  *strings.Builder
	TestFunc   func()                             `json:"-"`
	ResultFunc func(status *ExecutionStatus) bool `json:"-"`
}

// ResultRetCodePassedWhenSuccessful is used to mark a test case as PASSED when the command execution was successful (return code was 0)
func ResultRetCodePassedWhenSuccessful(status *ExecutionStatus) bool {
	return status.ExitStatus == 0
}

// ResultRetCodePassedWhenFailed is used to mark a test case as PASSED when the command execution failed (return code was not 0)
func ResultRetCodePassedWhenFailed(status *ExecutionStatus) bool {
	return status.ExitStatus != 0
}

func ResultStdoutNonEmptyPassed(status *ExecutionStatus) bool {
	return len(strings.Trim(status.Stdout, " ")) > 0
}

func ResultStdoutNonEmptyFailed(status *ExecutionStatus) bool {
	return len(strings.Trim(status.Stdout, " ")) == 0
}

// TODO:  is vague, correct it
func ResultRetCodeAndStdout(status *ExecutionStatus) bool {
	return status.ExitStatus == 0 && len(status.Stdout) > 0
}

var TestCases = []*TestCase{
	{"ENUM01", false, "PATH variables defined inside /etc", nil, []string{sh, grep, tr, cut, sort}, false, `for p in $(grep -ERh "^ *PATH=.*" /etc/ 2> /dev/null | tr -d '\\x22\\x27' | cut -d= -f2 | tr ":" "\n" | sort -u); do [ -d "$p" ] && echo "$p";done`, nil, nil, nil},
	{"ENUM02", false, "Writable files outside user's home (non-root users)", nil, []string{sh, find}, true, find_opts + `find / -path "$HOME" -prune -o $find_opts -not -type l -writable -print;[ "$(id -u)" != "0" ] && find  / -path "$HOME" -prune -o $find_opts -type l -uid $(id -u) -print`, nil, nil, nil},
	{"ENUM03", false, "Binaries with setuid bit", nil, []string{sh, find}, false, `find_opts='-path /proc -prune -o -path /sys -prune -o -path /dev -prune -o';find / $find_opts -perm -4000 -type f -print`, nil, nil, nil},
	{"ENUM04", false, "Binaries with setgid bit", nil, []string{sh, find}, false, `find_opts='-path /proc -prune -o -path /sys -prune -o -path /dev -prune -o';find / $find_opts -perm -2000 -type f -print`, nil, nil, nil},
	//{"CNPTST00", true, "It is forbidden to run a container as root", nil, nil, false, `[ $(id -u) != "0" ]`, nil, nil, ResultRetCodePassedWhenSuccessful},
	{"CNPTST01", true, "PATH variable defined inside /etc cannot contain '.'", []*Dependency{{Id: "ENUM01", Type: Stdout, VarName: "etc_exec_paths"}}, []string{sh, grep, tr}, false, `for ep in $etc_exec_paths; do [ "$ep" = "." ] && grep -ER "^ *PATH=.*" /etc/ 2> /dev/null | tr -d '\\x22\\x27' | grep -E "[=:]\.([:[:space:]]|\$)";done`, nil, nil, ResultStdoutNonEmptyFailed},
	{"CNPTST02", true, "User is forbidden to sudo without a password", nil, []string{sh, sudo}, true, `sudo -n true`, nil, nil, ResultRetCodePassedWhenFailed},
	{"CNPTST03", true, "User is forbidden to list sudo commands without a password", nil, []string{sh, sudo}, true, `sudo -nl`, nil, nil, ResultRetCodePassedWhenFailed},
	{"CNPTST04", true, "User is forbidden to read sudoers files (including /etc/sudoers.d/)", nil, []string{sh, grep}, true, `grep -R "" /etc/sudoers*`, nil, nil, ResultStdoutNonEmptyFailed},
	// TODO: fixed it - had echo printing directory
	{"CNPTST05", true, "User is forbidden to access other users home directories", nil, []string{sh}, true, `(for h in /home/*; do [ -d "$h" ] && [ "$h" != "$HOME" ] && (ls -la "$h/"); done)`, nil, nil, ResultStdoutNonEmptyFailed},
	{"CNPTST06", true, "Known exploitable binaries with setuid cannot be present on the system.", []*Dependency{{Id: "ENUM03", Type: TestFunc, VarName: "setuid_binaries"}}, nil, false, "", nil, nil, ResultStdoutNonEmptyFailed},
	{"CNPTST07", true, "Found setuid binaries cannot be writable by a user.", []*Dependency{{Id: "ENUM03", Type: Stdout, VarName: "setuid_binaries"}}, []string{sh}, true, `(for b in $setuid_binaries; do [ -x "$b" ] && [ -w "$b" ] && echo "$b" ;done)`, nil, nil, ResultStdoutNonEmptyFailed},
	{"CNPTST08", true, "Found setgid binaries cannot be writable by a user.", []*Dependency{{Id: "ENUM04", Type: Stdout, VarName: "setgid_binaries"}}, []string{sh}, true, `(for b in $setgid_binaries; do [ -x "$b" ] && [ -w "$b" ] && echo "$b" ;done)`, nil, nil, ResultStdoutNonEmptyFailed},
	{"CNPTST09", true, "Root directory (/root) cannot be readable by non-root users", []*Dependency{{Id: "ENUM04", Type: Stdout, VarName: "setgid_binaries"}}, []string{sh, ls}, true, `ls -ahl /root`, nil, nil, ResultRetCodePassedWhenFailed},
	{"CNPTST10", true, "git/svn repositories/directories cannot be present in a vm", nil, []string{sh, find}, false, find_opts + `find / $find_opts \( -name ".git" -o -name ".svn" \) -print`, nil, nil, ResultStdoutNonEmptyFailed},
	{"CNPTST11", true, "Critical files cannot be writable by a non-root user", []*Dependency{{Id: "ENUM02", Type: Stdout, VarName: "user_writable"}}, []string{sh}, true, "critical_writable=" + critical_writable + `;for uw in $user_writable; do [ -f "$uw" ] && IFS=$'\n'; for cw in ${critical_writable}; do [ "$cw" = "$uw" ] && [ -w "$cw" ] && ls -l $cw; done ; done`, nil, nil, ResultStdoutNonEmptyFailed},
	{"CNPTST12", true, "Critical directories cannot be writable by a non-root user", []*Dependency{{Id: "ENUM02", Type: Stdout, VarName: "user_writable"}}, []string{sh}, true, "critical_writable=" + critical_writable_dirs + `;for uw in $user_writable; do [ -d "$uw" ] && IFS=$'\n'; for cw in ${critical_writable_dirs}; do [ "$cw" = "$uw" ] && [ -w "$cw" ] && ls -ld $cw; done ; done`, nil, nil, ResultStdoutNonEmptyFailed},
	{"CNPTST13", true, "PATH directories cannot be writable by a non-root user", []*Dependency{{Id: "ENUM01", Type: Stdout, VarName: "exec_paths"}}, []string{sh}, true, `for ep in $exec_paths; do [ -d "$ep" ] && [ -w "$ep" ] && ls -ld "$ep"; done`, nil, nil, ResultStdoutNonEmptyFailed},
	{"CNPTST14", true, "History files cannot contain credentials", nil, []string{sh, grep}, false, `for h in .bash_history .history .histfile .zhistory; do [ -f "$HOME/$h" ] && grep  -Ei "(user|username|login|pass|password|pw|credentials)[=: ][a-z0-9]+" "$HOME/$h" | grep -v "systemctl" | grep -vE  "^find"; done`, nil, nil, ResultStdoutNonEmptyFailed},
	{"CNPTST15", true, "fstab/mtab files must not have credentials", nil, []string{sh, grep}, false, `grep $lse_grep_opts -Ei "(user|username|login|pass|password|pw|credentials|cred)[=:]" /etc/fstab /etc/mtab`, nil, nil, ResultStdoutNonEmptyFailed},
	// TODO: fix CNPTST16 - missing user check
	{"CNPTST16", true, "SSH files not owned by a current user must not be readable", nil, []string{sh, find}, true, find_opts + `find / $find_opts \( -name "*id_dsa*" -o -name "*id_rsa*" -o -name "*id_ecdsa*" -o -name "*id_ed25519*" -o -name "known_hosts" -o -name "authorized_hosts" -o -name "authorized_keys" \) -readable ! -user $USER -exec ls -la {} \;`, nil, nil, ResultStdoutNonEmptyFailed},
	{"CNPTST17", true, "/etc/passwd must not have hashes", nil, []string{sh, grep}, false, `grep -v "^[^:]*:[x]" /etc/passwd`, nil, nil, ResultStdoutNonEmptyFailed},
	{"CNPTST18", true, "/etc/group must not have hashes", nil, []string{sh, grep}, false, `grep -v "^[^:]*:[x]" /etc/group`, nil, nil, ResultStdoutNonEmptyFailed},
	{"CNPTST19", true, "shadow files must not be readable for non-root user", nil, []string{sh}, true, `for sf in "shadow" "shadow-" "shadow~" "gshadow" "gshadow-" "master.passwd"; do [ -r "/etc/$sf" ] && printf "%s\n---\n" "/etc/$sf" && cat "/etc/$sf" && printf "\n\n";done`, nil, nil, ResultStdoutNonEmptyFailed},
	{"CNPTST20", true, "root must not be able to log in via SSH", nil, []string{sh, grep}, false, `grep -E "^[[:space:]]*PermitRootLogin " /etc/ssh/sshd_config | grep -E "(yes|without-password|prohibit-password)"`, nil, nil, ResultRetCodePassedWhenFailed},
	// TODO: fixed since there was check for root
	{"CNPTST21", true, "binaries with caps must not be writable by a non-root user", nil, []string{sh, getcap}, true, `cap_bin=$(getcap -r / 2>/dev/null);(for b in $(printf "$cap_bin\n" | cut -d" " -f1); do [ -w "$b" ] && echo "$b"; done)`, nil, nil, ResultStdoutNonEmptyFailed},
	{"CNPTST22", true, "A binary must not have all caps assigned", nil, []string{sh, grep, getcap}, false, `cap_bin=$(getcap -r / 2>/dev/null);printf "$cap_bin\n" | grep -v "cap_"`, nil, nil, ResultStdoutNonEmptyFailed},
	{"CNPTST23", true, "A user must not have capabilities assigned", nil, []string{sh, grep}, true, `user_caps=$(grep -v "^#\|none\|^$" /etc/security/capability.conf 2>/dev/null);printf "$user_caps\n" | grep "$USER"`, nil, nil, ResultStdoutNonEmptyFailed},
	{"CNPTST24", true, "Cron tasks must not be writable by a non-root user", nil, []string{sh, find}, true, `find -L /etc/cron* /etc/anacron /var/spool/cron -writable 2>/dev/null`, nil, nil, ResultStdoutNonEmptyFailed},
	//{"CNPTST25", true, "A non-root user must not read other users crontabs", nil, []string{ls}, `[ "$(id -u)" != "0" ] && { ls -la /var/spool/cron/crontabs/*; }`, nil, ResultStdout},
	{"CNPTST25", true, "A non-root user must not read other users crontabs", nil, []string{sh, id, ls}, true, `for h in /var/spool/cron/crontabs/*; do [ -r "$h" ] && (cat "$h"); done`, nil, nil, ResultStdoutNonEmptyFailed},
	{"CNPTST26", true, "A non-root user must not be able to list other users cron tasks", nil, []string{sh, "/usr/bin/crontab"}, true, `[ "$(id -u)" != "0" ] && (for u in $(cut -d: -f 1 /etc/passwd); do [ "$u" != "$(id -un)" ] && crontab -l -u "$u"; done 2>/dev/null)`, nil, nil, ResultStdoutNonEmptyFailed},
	{"CNPTST27", true, "Any paths present in cron jobs must not be writable by a non-root user", nil, []string{sh, grep, sort}, true, `for p in $(grep --color=never -hERoi "/[a-z0-9_/\.\-]+" /etc/cron* | grep -Ev "/dev/(null|zero|random|urandom)" | sort -u); do [ -w "$p" ] && echo "$p"; done 2>/dev/null`, nil, nil, ResultStdoutNonEmptyFailed},
	{"CNPTST28", true, "Executable paths present in cron jobs must not be writable by a non-root user", []*Dependency{{Id: "CNPTST27", Type: Stdout, VarName: "user_writable_cron_paths"}}, []string{sh, grep}, true, `for path in $user_writable_cron_paths; do [ -w "$path" ] && [ -x "$path" ] && grep  -R "$path" /etc/crontab /etc/cron.d/ /etc/anacrontab ; done 2>/dev/null`, nil, nil, ResultStdoutNonEmptyFailed},
	{"CNPTST29", true, "A non-root user must not be able to write to any system timer", []*Dependency{{Id: "ENUM02", Type: Stdout, VarName: "user_writable"}}, []string{sh}, true, `printf "$user_writable\n" | grep -E "\.timer$" 2>/dev/null`, nil, nil, ResultStdoutNonEmptyFailed},
	{"CNPTST30", true, "A non-root user must not be able to write to any service unit file", []*Dependency{{Id: "ENUM02", Type: Stdout, VarName: "user_writable"}}, []string{sh, grep}, true, `printf "$user_writable\n" | grep -E "^/etc/(init/|init\.d/|rc\.d/|rc[0-9S]\.d/|rc\.local|inetd\.conf|xinetd\.conf|xinetd\.d/)"`, nil, nil, ResultStdoutNonEmptyFailed},
	{"CNPTST31", true, "A non-root user must not be able to write to any executable used by service", nil, []string{sh, grep, tr, sort}, true, `for b in $(grep -ERvh "^#" /etc/inetd.conf /etc/xinetd.conf /etc/xinetd.d/ /etc/init.d/ /etc/rc* 2>/dev/null | tr -s "[[:space:]]" "\n" | grep -E "^/" | grep -Ev "^/(dev|run|sys|proc|tmp)/" | sort -u); do [ -x "$b" ] && [ -w "$b" ] && echo "$b" ;done`, nil, nil, ResultStdoutNonEmptyFailed},
	{"CNPTST32", true, "A non-root user must not be able to write to any systemd files", []*Dependency{{Id: "ENUM02", Type: Stdout, VarName: "user_writable"}}, []string{sh, grep}, true, `printf "$user_writable\n" | grep -E "^/(etc/systemd/|lib/systemd/).+\.service$" 2>/dev/null`, nil, nil, ResultStdoutNonEmptyFailed},
	{"CNPTST33", true, "A non-root user must not be able to write to any executable used by systemd services", nil, []string{sh, grep, tr, sort}, true, `for b in $(grep -ERh "^Exec" /etc/systemd/ /lib/systemd/ 2>/dev/null | tr "=" "\n" | tr -s "[[:space:]]" "\n" | grep -E "^/" | grep -Ev "^/(dev|run|sys|proc|tmp)/" | sort -u); do [ -x "$b" ] && [ -w "$b" ] && echo "$b" ;done 2>/dev/null`, nil, nil, ResultStdoutNonEmptyFailed},
	{"CNPTST34", true, "A non-root user must not own systemd files", nil, []string{sh, find, ls}, true, `find /lib/systemd/ /etc/systemd \! -uid 0 -type f -exec ls -la {} \; 2>/dev/null`, nil, nil, ResultStdoutNonEmptyFailed},
	{"CNPTST35", true, "mysql root account must be password protected", nil, []string{sh, "mysqladmin"}, false, `mysqladmin -uroot version`, nil, nil, ResultRetCodePassedWhenFailed},
	{"CNPTST36", true, "mysql root account is forbidden to use 'root' as a password", nil, []string{sh, "mysqladmin"}, false, `mysqladmin -uroot -proot version 2>/dev/null`, nil, nil, ResultRetCodePassedWhenFailed},
	{"CNPTST37", true, "$HOME/.mysql_history file must not contain credentials. Link it to /dev/null", nil, []string{sh, grep}, false, `grep -Ei "(pass|identified by|md5\()" "$HOME/.mysql_history" 2>/dev/null`, nil, nil, ResultStdoutNonEmptyFailed},
	{"CNPTST38", true, "postgres templates must be password protected", nil, []string{sh, "psql", grep}, false, `(psql -U postgres template0 -c "select version()" 2>/dev/null | grep version) || (psql -U postgres template1 -c "select version()" 2>/dev/null | grep version) || (psql -U pgsql template0 -c "select version()" 2>/dev/null | grep version) || (psql -U pgsql template1 -c "select version()" 2>/dev/null | grep version)`, nil, nil, ResultRetCodePassedWhenFailed},
	{"CNPTST39", true, "mongodb must be password protected", nil, []string{sh, "mongo", grep}, false, `echo "show dbs" | mongo --quiet | grep -E "(admin|config|local)"`, nil, nil, ResultStdoutNonEmptyFailed},
	// {"CNPTST40", true, "docker must not be available in a vm", nil, []string{sh, "docker"}, false, `docker --version; docker ps -a; docker images`, nil, nil, ResultStdout},
	{"CNPTST40", true, "a non-root user must not be a member of docker group", nil, []string{sh, grep, groups}, true, `groups | grep -o docker`, nil, nil, ResultStdoutNonEmptyFailed},
	{"CNPTST41", true, "a non-root user must not be a member of a lxc or lxd group", nil, []string{sh, grep, groups}, true, `groups | grep  "lxc\|lxd"`, nil, nil, ResultStdoutNonEmptyFailed},
	//{"CNPTST16", true, "", nil, []string{grep}, ``, nil, ResultStdout},
}

// TODO: consider using this variables directly in a command by concatenating them into the command instead of a shell variable,
//
//	it would help to limit dependency on the shell.
var find_opts = `find_opts='-path /proc -prune -o -path /sys -prune -o -path /dev -prune -o -path /run -prune -o';`

var CommonSetuidBinaries = []string{
	"/bin/fusermount",
	"/bin/mount",
	"/bin/ntfs-3g",
	"/bin/ping",
	"/bin/ping6",
	"/bin/su",
	"/bin/umount",
	"/lib64/dbus-1/dbus-daemon-launch-helper",
	"/sbin/mount.ecryptfs_private",
	"/sbin/mount.nfs",
	"/sbin/pam_timestamp_check",
	"/sbin/pccardctl",
	"/sbin/unix2_chkpwd",
	"/sbin/unix_chkpwd",
	"/usr/bin/Xorg",
	"/usr/bin/arping",
	"/usr/bin/at",
	"/usr/bin/beep",
	"/usr/bin/chage",
	"/usr/bin/chfn",
	"/usr/bin/chsh",
	"/usr/bin/crontab",
	"/usr/bin/expiry",
	"/usr/bin/firejail",
	"/usr/bin/fusermount",
	"/usr/bin/fusermount-glusterfs",
	"/usr/bin/fusermount3",
	"/usr/bin/gpasswd",
	"/usr/bin/kismet_capture",
	"/usr/bin/mount",
	"/usr/bin/mtr",
	"/usr/bin/newgidmap",
	"/usr/bin/newgrp",
	"/usr/bin/newuidmap",
	"/usr/bin/ntfs-3g",
	"/usr/bin/passwd",
	"/usr/bin/pkexec",
	"/usr/bin/pmount",
	"/usr/bin/procmail",
	"/usr/bin/pumount",
	"/usr/bin/staprun",
	"/usr/bin/su",
	"/usr/bin/sudo",
	"/usr/bin/sudoedit",
	"/usr/bin/traceroute6.iputils",
	"/usr/bin/umount",
	"/usr/bin/weston-launch",
	"/usr/lib/chromium-browser/chrome-sandbox",
	"/usr/lib/dbus-1.0/dbus-daemon-launch-helper",
	"/usr/lib/dbus-1/dbus-daemon-launch-helper",
	"/usr/lib/eject/dmcrypt-get-device",
	"/usr/lib/openssh/ssh-keysign",
	"/usr/lib/policykit-1/polkit-agent-helper-1",
	"/usr/lib/polkit-1/polkit-agent-helper-1",
	"/usr/lib/pt_chown",
	"/usr/lib/snapd/snap-confine",
	"/usr/lib/spice-gtk/spice-client-glib-usb-acl-helper",
	"/usr/lib/x86_64-linux-gnu/lxc/lxc-user-nic",
	"/usr/lib/xorg/Xorg.wrap",
	"/usr/libexec/Xorg.wrap",
	"/usr/libexec/abrt-action-install-debuginfo-to-abrt-cache",
	"/usr/libexec/cockpit-session",
	"/usr/libexec/dbus-1/dbus-daemon-launch-helper",
	"/usr/libexec/gstreamer-1.0/gst-ptp-helper",
	"/usr/libexec/openssh/ssh-keysign",
	"/usr/libexec/polkit-1/polkit-agent-helper-1",
	"/usr/libexec/polkit-agent-helper-1",
	"/usr/libexec/pt_chown",
	"/usr/libexec/qemu-bridge-helper",
	"/usr/libexec/spice-client-glib-usb-acl-helper",
	"/usr/libexec/spice-gtk-x86_64/spice-client-glib-usb-acl-helper",
	"/usr/local/share/panasonic/printer/bin/L_H0JDUCZAZ",
	"/usr/sbin/exim4",
	"/usr/sbin/grub2-set-bootflag",
	"/usr/sbin/mount.nfs",
	"/usr/sbin/mtr-packet",
	"/usr/sbin/pam_timestamp_check",
	"/usr/sbin/pppd",
	"/usr/sbin/pppoe-wrapper",
	"/usr/sbin/suexec",
	"/usr/sbin/unix_chkpwd",
	"/usr/sbin/userhelper",
	"/usr/sbin/usernetctl",
	"/usr/sbin/uuidd",
}

var ExploitableSuidBinaries = []string{
	"aa-exec",
	"ab",
	"agetty",
	"alpine",
	"ar",
	"arj",
	"arp",
	"as",
	"ascii-xfr",
	"ash",
	"aspell",
	"atobm",
	"awk",
	"base32",
	"base64",
	"basenc",
	"basez",
	"bash",
	"bc",
	"Binary",
	"bridge",
	"busybox",
	"bzip2",
	"cabal",
	"capsh",
	"cat",
	"chmod",
	"choom",
	"chown",
	"chroot",
	"cmp",
	"column",
	"comm",
	"cp",
	"cpio",
	"cpulimit",
	"csh",
	"csplit",
	"csvtool",
	"cupsfilter",
	"curl",
	"cut",
	"dash",
	"date",
	"dd",
	"debugfs",
	"dialog",
	"diff",
	"dig",
	"distcc",
	"dmsetup",
	"docker",
	"dosbox",
	"ed",
	"efax",
	"elvish",
	"emacs",
	"env",
	"eqn",
	"espeak",
	"expand",
	"expect",
	"file",
	"find",
	"fish",
	"flock",
	"fmt",
	"fold",
	"gawk",
	"gcore",
	"gdb",
	"genie",
	"genisoimage",
	"gimp",
	"grep",
	"gtester",
	"gzip",
	"hd",
	"head",
	"hexdump",
	"highlight",
	"hping3",
	"iconv",
	"install",
	"ionice",
	"ip",
	"ispell",
	"jjs",
	"join",
	"jq",
	"jrunscript",
	"julia",
	"ksh",
	"ksshell",
	"kubectl",
	"ld.so",
	"less",
	"logsave",
	"look",
	"lua",
	"make",
	"mawk",
	"more",
	"mosquitto",
	"msgattrib",
	"msgcat",
	"msgconv",
	"msgfilter",
	"msgmerge",
	"msguniq",
	"multitime",
	"mv",
	"nasm",
	"nawk",
	"ncftp",
	"nft",
	"nice",
	"nl",
	"nm",
	"nmap",
	"node",
	"nohup",
	"od",
	"openssl",
	"openvpn",
	"pandoc",
	"paste",
	"perf",
	"perl",
	"pexec",
	"pg",
	"php",
	"pidstat",
	"pr",
	"ptx",
	"python",
	"rc",
	"readelf",
	"restic",
	"rev",
	"rlwrap",
	"rsync",
	"rtorrent",
	"run-parts",
	"rview",
	"rvim",
	"sash",
	"scanmem",
	"sed",
	"setarch",
	"setfacl",
	"setlock",
	"shuf",
	"soelim",
	"softlimit",
	"sort",
	"sqlite3",
	"ss",
	"ssh-agent",
	"ssh-keygen",
	"ssh-keyscan",
	"sshpass",
	"start-stop-daemon",
	"stdbuf",
	"strace",
	"strings",
	"sysctl",
	"systemctl",
	"tac",
	"tail",
	"taskset",
	"tbl",
	"tclsh",
	"tee",
	"tftp",
	"tic",
	"time",
	"timeout",
	"troff",
	"ul",
	"unexpand",
	"uniq",
	"unshare",
	"unzip",
	"update-alternatives",
	"uudecode",
	"uuencode",
	"vagrant",
	"view",
	"vigr",
	"vim",
	"vimdiff",
	"vipw",
	"w3m",
	"watch",
	"wc",
	"wget",
	"whiptail",
	"xargs",
	"xdotool",
	"xmodmap",
	"xmore",
	"xxd",
	"xz",
	"yash",
	"zsh",
	"zsoelim",
}

var critical_writable string = `"
/etc/apache2/apache2.conf
/etc/apache2/httpd.conf
/etc/bash.bashrc
/etc/bash_completion
/etc/bash_completion.d/*
/etc/environment
/etc/environment.d/*
/etc/hosts.allow
/etc/hosts.deny
/etc/httpd/conf/httpd.conf
/etc/httpd/httpd.conf
/etc/incron.conf
/etc/incron.d/*
/etc/logrotate.d/*
/etc/modprobe.d/*
/etc/pam.d/*
/etc/passwd
/etc/php*/fpm/pool.d/*
/etc/php/*/fpm/pool.d/*
/etc/profile
/etc/profile.d/*
/etc/rc*.d/*
/etc/rsyslog.d/*
/etc/shadow
/etc/skel/*
/etc/sudoers
/etc/sudoers.d/*
/etc/supervisor/conf.d/*
/etc/supervisor/supervisord.conf
/etc/sysctl.conf
/etc/sysctl.d/*
/etc/uwsgi/apps-enabled/*
/root/.ssh/authorized_keys
"`

var critical_writable_dirs string = `"
/etc/bash_completion.d
/etc/cron.d
/etc/cron.daily
/etc/cron.hourly
/etc/cron.weekly
/etc/environment.d
/etc/logrotate.d
/etc/modprobe.d
/etc/pam.d
/etc/profile.d
/etc/rsyslog.d/
/etc/sudoers.d/
/etc/sysctl.d
/root
"`
