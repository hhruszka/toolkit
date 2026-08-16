package testengine

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"
	"unicode"
)

// Logic:
// - each entry can be either a test case or an enumeration.
// - each entry can have number of prerequisites that are executed recurrently
// - prerequisite can only state false or true or provide stdout
// - only test cases marked as "not test" case can be on a list of dependencies
// - test case decides about its result, prerequisite delivers only information (data, or execution result)
// - if a test case cannot be executed due to missing binary on the SUT then such a test case is marked as passed.
// In the future the latter behaviour might be changed/modified by introducing configuration parameters that would decide
// about a result (failed, passed) in such cases.

const (
	cut    = "cut"
	env    = "env"
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
	NonRoot bool   // if true need to be run by non-root user
	Command string `json:"Command"`
	//stdOutVar  *strings.Builder
	TestFunc   func(testId string, depExecResults map[string]map[string]*ExecutionStatus) *ExecutionStatus `json:"-"`
	ResultFunc func(status *ExecutionStatus) bool                                                          `json:"-"`
}

func isEmpty(str ...string) bool {
	trim := func(s string) string {
		return strings.TrimFunc(s, func(r rune) bool {
			return unicode.IsSpace(r)
		})
	}

	if len(str) == 0 {
		return true
	}
	for _, line := range str {
		if len(trim(line)) > 0 {
			return false
		}
	}
	return true
}

func PrintTestCasesOnly() *bytes.Buffer {
	var buf bytes.Buffer

	w := tabwriter.NewWriter(&buf, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "Test Id\tAbstract")
	for _, test := range TestCases {
		if test.IsTest {
			fmt.Fprintf(w, "%s\t%s\n", test.Id, test.Abstract)
		}
	}
	w.Flush()

	return &buf
}

// PassedWhenRetCodeSuccessful is used to mark a test case as PASSED when the command execution was successful (return code was 0)
func PassedWhenRetCodeSuccessful(status *ExecutionStatus) bool {
	return status.RetCode == Success
}

// PassedWhenRetCodeFailed is used to mark a test case as PASSED when the command execution failed (return code was not 0)
func PassedWhenRetCodeFailed(status *ExecutionStatus) bool {
	return status.RetCode != Success
}

// PassedWhenStdoutNonEmpty fails the test case when Stdout is empty
func PassedWhenStdoutNonEmpty(status *ExecutionStatus) bool {
	return isEmpty(status.Stdout...) == false
}

// PassedWhenStdoutEmpty passed when the test case when Stdout is empty
func PassedWhenStdoutEmpty(status *ExecutionStatus) bool {
	return isEmpty(status.Stdout...) == true
}

// PassedWhenRetCodeSuccessfulAndStdoutNonEmpty
func PassedWhenRetCodeSuccessfulAndStdoutNonEmpty(status *ExecutionStatus) bool {
	return status.RetCode == Success && isEmpty(status.Stdout...) == false
}

var TestCases = []*TestCase{
	{"ENUM01", false, "PATH variables defined inside /etc", nil, []string{sh, grep, tr, cut, sort}, false, `for p in $(grep -ERh "^ *PATH=.*" /etc/ 2> /dev/null | tr -d $(printf '\x22\x27') | cut -d= -f2 | tr ":" "\n" | sort -u); do [ -d "$p" ] && echo "$p";done`, nil, nil},
	{"ENUM02", false, "Writable files outside user's home (non-root users)", nil, []string{sh, find}, true, find_opts + `find / -path "$HOME" -prune -o $find_opts \( -writable -o \( ! -type l -uid $(id -u) \) \) -print`, nil, nil},
	{"ENUM03", false, "Binaries with setuid bit owned by root", nil, []string{sh, find}, false, `find_opts='-path /proc -prune -o -path /sys -prune -o -path /dev -prune -o';find / $find_opts -perm -4000 -type f -user 0 -print`, nil, nil},
	{"ENUM04", false, "Binaries with setgid bit owned by root", nil, []string{sh, find}, false, `find_opts='-path /proc -prune -o -path /sys -prune -o -path /dev -prune -o';find / $find_opts -perm -2000 -type f -user 0 -print`, nil, nil},
	{"ENUM05", false, "Processes", nil, nil, false, `q`, enum05, nil},
	// # Get the user's default shell and run interactively
	//sudo -u username $(getent passwd username | cut -d: -f7) -i -c 'env'
	{"ENUM06", false, "Environment variables", nil, []string{env}, false, `env`, nil, nil},
	{"ENUM07", false, "sudo output", nil, []string{sudo}, false, `sudo -nl`, nil, nil},
	//{"VNFBPT00", true, "It is forbidden to run a container as root", nil, nil, false, "", cnpt00, PassedWhenRetCodeFailed},
	{"VNFBPT01", true, "PATH variable defined inside /etc cannot contain '.'", []*Dependency{{Id: "ENUM01", Type: Stdout, VarName: "etc_exec_paths"}}, []string{sh, grep, tr}, false, `for ep in $etc_exec_paths; do [ "$ep" = "." ] && grep -ER "^ *PATH=.*" /etc/ 2> /dev/null | tr -d $(printf '\x22\x27') | grep -E "[=:]\./*([:[:space:]]|\$)";done`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT02", true, "User is forbidden to sudo without a password", nil, []string{sh, sudo}, true, `sudo -n true`, nil, PassedWhenRetCodeFailed},
	{"VNFBPT03", true, "User is forbidden to list sudo commands without a password", nil, []string{sh, sudo}, true, `sudo -nl`, nil, PassedWhenRetCodeFailed},
	{"VNFBPT04", true, "User is forbidden to read sudoers files (including /etc/sudoers.d/)", nil, []string{sh, grep}, true, `grep -R "" /etc/sudoers*`, nil, PassedWhenStdoutEmpty},
	// TODO: fixed it - had echo printing directory
	{"VNFBPT05", true, "User is forbidden to access other users home directories.", nil, []string{sh}, true, `[ "$(echo /home/*)" != '/home/*' ] && (for h in /home/*; do [ -d "$h" ] && (echo $HOME | grep -q -o $h) || ([ -r $h ] && [ -x $h ] && { echo $h;ls -l $h ;}); done)`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT06", true, "Known exploitable setuid binaries cannot be present on the system.", []*Dependency{{Id: "ENUM03", Type: TestFunc, VarName: "setuid_binaries"}}, nil, false, "", vnfbpt06, PassedWhenStdoutEmpty},
	{"VNFBPT07", true, "Found setuid binaries cannot be writable by a user.", []*Dependency{{Id: "ENUM03", Type: Stdout, VarName: "setuid_binaries"}}, []string{sh}, true, `(for b in $setuid_binaries; do [ -x "$b" ] && [ -w "$b" ] && echo "$b" ;done)`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT08", true, "Found setgid binaries cannot be writable by a user.", []*Dependency{{Id: "ENUM04", Type: Stdout, VarName: "setgid_binaries"}}, []string{sh}, true, `(for b in $setgid_binaries; do [ -x "$b" ] && [ -w "$b" ] && echo "$b" ;done)`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT09", true, "Root directory (/root) cannot be readable by non-root users", nil, []string{sh, ls}, true, `test -r /root && test -x /root`, nil, PassedWhenRetCodeFailed},
	{"VNFBPT10", true, "git/svn repositories/directories cannot be present in a vm", nil, []string{sh, find}, false, find_opts + `find / $find_opts \( -name ".git" -o -name ".svn" \) -print`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT11", true, "Critical files cannot be writable by a non-root user", []*Dependency{{Id: "ENUM02", Type: Stdout, VarName: "user_writable"}}, []string{sh}, true, "critical_writable=" + critical_writable + `;IFS=$'\n';for uw in $user_writable; do [ -f "$uw" ] || continue; for cw in ${critical_writable}; do [ "$cw" = "$uw" ] && ls -l "$cw"; done; done`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT12", true, "Critical directories cannot be writable by non-root users", []*Dependency{{Id: "ENUM02", Type: Stdout, VarName: "user_writable"}}, []string{sh}, true, "critical_writable_dirs=" + critical_writable_dirs + `;IFS=$'\n';for uw in $user_writable; do [ -d "$uw" ] || continue; for cw in ${critical_writable_dirs}; do [ "$cw" = "$uw" ] && ls -ld "$cw"; done; done`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT13", true, "PATH directories cannot be writable by a non-root user", []*Dependency{{Id: "ENUM01", Type: Stdout, VarName: "exec_paths"}}, []string{sh}, true, `for ep in $exec_paths; do [ -d "$ep" ] && [ -w "$ep" ] && ls -ld "$ep"; done`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT14", true, "History files cannot contain credentials", nil, []string{sh, grep}, false, `for h in .bash_history .history .histfile .zhistory; do [ -f "$HOME/$h" ] && grep  -Ei "(user|username|login|pass|password|pw|credentials)[=: ][a-z0-9]+" "$HOME/$h" | grep -v "systemctl" | grep -vE  "^find"; done`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT15", true, "fstab/mtab files must not have credentials", nil, []string{sh, grep}, false, `grep $lse_grep_opts -Ei "(user|username|login|pass|password|pw|credentials|cred)[=:]" /etc/fstab /etc/mtab`, nil, PassedWhenStdoutEmpty},
	// TODO: fix VNFBPT16 - missing user check, add uid shell variable containing user id and use it instead of $USER shell variable or use $(id -u) instead of.
	{"VNFBPT16", true, "SSH files not owned by the current user must not be readable", nil, []string{sh, find}, true, find_opts + `find / $find_opts \( -name "*id_dsa*" -o -name "*id_rsa*" -o -name "*id_ecdsa*" -o -name "*id_ed25519*" -o -name "known_hosts" -o -name "authorized_hosts" -o -name "authorized_keys" \) -readable ! -user $USER -exec ls -la {} \; 2>/dev/null`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT17", true, "/etc/passwd must not have password hashes", nil, []string{sh, grep}, false, `grep -v "^[^:]*:[x]" /etc/passwd`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT18", true, "/etc/group must not have password hashes", nil, []string{sh, grep}, false, `grep -v "^[^:]*:[x]" /etc/group`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT19", true, "shadow files must not be readable for non-root user", nil, []string{sh}, true, `for sf in "shadow" "shadow-" "shadow~" "gshadow" "gshadow-" "master.passwd"; do [ -r "/etc/$sf" ] && printf "%s\n---\n" "/etc/$sf" && cat "/etc/$sf" && printf "\n\n";done`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT20", true, "root must not be able to log in via SSH", nil, []string{sh, grep}, false, `grep -E "^[[:space:]]*PermitRootLogin " /etc/ssh/sshd_config | grep -E "(yes|without-password|prohibit-password)"`, nil, PassedWhenRetCodeFailed},
	// TODO: fixed since there was check for root
	// New version:  if [ ! -w ./test.txt ];then  [ -O ./test.txt ] && getcap ./text.txt;ls -l ./text.txt;else getcap ./file.txt;ls -l ./text.txt;fi
	//{"VNFBPT21", true, "binaries with caps must not be writable by a non-root user", nil, []string{sh, getcap, ls}, true, `cap_bin=$(getcap -r / 2>/dev/null);(for b in $(printf "$cap_bin\n" | cut -d" " -f1); do [ -w "$b" ] && ls -l "$b"; done)`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT21", true, "Binaries with caps must not be writable by non-root users", nil, []string{sh, getcap, ls}, true, `cap_bin=$(getcap -r / 2>/dev/null);(for b in $(printf "$cap_bin\n" | cut -d" " -f1); do [ -w $b ] && (getcap $b;ls -l $b); done)`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT22", true, "A binary must not have all caps assigned", nil, []string{sh, grep, getcap}, false, `cap_bin=$(getcap -r / 2>/dev/null);printf "$cap_bin\n" | grep -v "cap_"`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT23", true, "A user must not have capabilities assigned", nil, []string{sh, grep}, true, `user_caps=$(grep -v "^#\|none\|^$" /etc/security/capability.conf 2>/dev/null);printf "$user_caps\n" | grep "$USER"`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT24", true, "Cron tasks must not be writable by non-root users", nil, []string{sh, find}, true, `find -L /etc/cron* /etc/anacron /var/spool/cron -writable 2>/dev/null`, nil, PassedWhenStdoutEmpty},
	//{"VNFBPT25", true, "A non-root user must not read other users crontabs", nil, []string{ls}, `[ "$(id -u)" != "0" ] && { ls -la /var/spool/cron/crontabs/*; }`, nil, ResultStdout},
	{"VNFBPT25", true, "A non-root user must not read other users crontabs", nil, []string{sh, id, ls}, true, `for h in /var/spool/cron/crontabs/*; do [ -r "$h" ] && (cat "$h"); done`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT26", true, "A non-root user must not be able to list other users cron tasks", nil, []string{sh, "/usr/bin/crontab"}, true, `(for u in $(cut -d: -f 1 /etc/passwd); do [ "$u" != "$(id -un)" ] && crontab -l -u "$u"; done 2>/dev/null)`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT27", true, "Any paths present in cron jobs must not be writable by a non-root user", nil, []string{sh, grep, sort}, true, `for p in $(grep --color=never -hERoi "/[a-z0-9_/\.\-]+" /etc/cron* | grep -Ev "/dev/(null|zero|random|urandom)" | sort -u); do [ -w "$p" ] && echo "$p"; done 2>/dev/null`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT28", true, "Executable paths present in cron jobs must not be writable by a non-root user", []*Dependency{{Id: "VNFBPT27", Type: Stdout, VarName: "user_writable_cron_paths"}}, []string{sh, grep}, true, `for path in $user_writable_cron_paths; do [ -w "$path" ] && [ -x "$path" ] && grep  -R "$path" /etc/crontab /etc/cron.d/ /etc/anacrontab ; done 2>/dev/null`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT29", true, "A non-root user must not be able to write to any system timer", []*Dependency{{Id: "ENUM02", Type: Stdout, VarName: "user_writable"}}, []string{sh}, true, `printf "$user_writable\n" | grep -E "\.timer$" 2>/dev/null`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT30", true, "A non-root user must not be able to write to any legacy init configuration", []*Dependency{{Id: "ENUM02", Type: Stdout, VarName: "user_writable"}}, []string{sh, grep}, true, `printf "$user_writable\n" | grep -E "^/etc/(init/|init\.d/|rc\.d/|rc[0-9S]\.d/|rc\.local|inetd\.conf|xinetd\.conf|xinetd\.d/)"`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT31", true, "A non-root user must not be able to write to any executable used by legacy init", nil, []string{sh, grep, tr, sort}, true, `for b in $(grep -ERvh "^#" /etc/inetd.conf /etc/xinetd.conf /etc/xinetd.d/ /etc/init.d/ /etc/rc* 2>/dev/null | tr -s "[[:space:]]" "\n" | grep -E "^/" | grep -Ev "^/(dev|run|sys|proc|tmp)(/|$)" | sort -u); do [ -x "$b" ] && [ -w "$b" ] && echo "$b" ;done`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT32", true, "A non-root user must not be able to write to any systemd unit files or drop-ins", []*Dependency{{Id: "ENUM02", Type: Stdout, VarName: "user_writable"}}, []string{sh, grep}, true, `printf "$user_writable\n" | grep -E "^/(etc/systemd/|lib/systemd/).+\.service$" 2>/dev/null`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT33", true, "A non-root user must not be able to write to any executable used by systemd services", nil, []string{sh, grep, tr, sort}, true, `for b in $(grep -ERh "^Exec" /etc/systemd/ /lib/systemd/ 2>/dev/null | tr "=" "\n" | tr -s "[[:space:]]" "\n" | grep -E "^/" | grep -Ev "^/(dev|run|sys|proc|tmp)/" | sort -u); do [ -f $b ] && [ -x "$b" ] && [ -w "$b" ] && echo "$b" ;done 2>/dev/null`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT34", true, "A non-root user must not own systemd files", nil, []string{sh, find, ls}, false, `find /lib/systemd/ /etc/systemd ! -uid 0 -type f -exec ls -la {} \; 2>/dev/null`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT35", true, "mysql root account must be password protected", nil, []string{sh, "mysqladmin"}, false, `mysqladmin -uroot version`, nil, PassedWhenRetCodeFailed},
	{"VNFBPT36", true, "mysql root account is forbidden to use 'root' as a password", nil, []string{sh, "mysqladmin"}, false, `mysqladmin -uroot -proot version 2>/dev/null`, nil, PassedWhenRetCodeFailed},
	{"VNFBPT37", true, "$HOME/.mysql_history file must not contain credentials.", nil, []string{sh, grep}, false, `grep -Ei "(pass|identified by|md5\()" "$HOME/.mysql_history" 2>/dev/null`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT38", true, "postgres templates must be password protected", nil, []string{sh, "psql", grep}, false, `(psql -U postgres template0 -c "select version()" 2>/dev/null | grep version) || (psql -U postgres template1 -c "select version()" 2>/dev/null | grep version) || (psql -U pgsql template0 -c "select version()" 2>/dev/null | grep version) || (psql -U pgsql template1 -c "select version()" 2>/dev/null | grep version)`, nil, PassedWhenRetCodeFailed},
	{"VNFBPT39", true, "mongodb must be password protected", nil, []string{sh, "mongo", grep}, false, `echo "show dbs" | mongo --quiet | grep -E "(admin|config|local)"`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT40", true, "no_all_squash is forbidden in /etc/exports", nil, []string{sh, grep}, false, `grep "no_all_squash" /etc/exports`, nil, PassedWhenRetCodeFailed},
	{"VNFBPT41", true, "no_root_squash is forbidden in /etc/exports", nil, []string{sh, grep}, false, `grep "no_root_squash" /etc/exports`, nil, PassedWhenRetCodeFailed},
	{"VNFBPT42", true, "A non-root user must not be a member of docker group", nil, []string{sh, grep, groups}, true, `groups | grep -o docker`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT43", true, "A non-root user must not be a member of a lxc or lxd group", nil, []string{sh, grep, groups}, true, `groups | grep -E "lxc|lxd"`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT44", true, "Processes run by root must not be writable by other users", []*Dependency{{Id: "ENUM05", Type: TestFunc, VarName: ""}}, []string{sh, grep, groups}, true, ``, vnfbpt44, PassedWhenRetCodeSuccessful},
	{"VNFBPT45", true, "Processes run by root must not be invoked with files readable by other users", []*Dependency{{Id: "ENUM05", Type: TestFunc, VarName: ""}}, []string{sh, grep, groups}, true, ``, vnfbpt45, PassedWhenRetCodeSuccessful},
	{"VNFBPT46", true, "Processes run by root must not be invoked with files writable by other users", []*Dependency{{Id: "ENUM05", Type: TestFunc, VarName: ""}}, []string{sh, grep, groups}, true, ``, vnfbpt46, PassedWhenRetCodeSuccessful},
	{"VNFBPT47", true, "Processes must not be invoked with files containing secrets", []*Dependency{{Id: "ENUM05", Type: TestFunc, VarName: ""}}, []string{sh, grep, groups}, true, ``, vnfbpt47, PassedWhenRetCodeSuccessful},
	{"VNFBPT48", true, "Processes must not be invoked with secrets/credentials on the command line", []*Dependency{{Id: "ENUM05", Type: TestFunc, VarName: ""}}, []string{sh, grep, groups}, true, ``, vnfbpt48, PassedWhenRetCodeSuccessful},
	{"VNFBPT49", true, "A non-root user must not have read permissions to the audit log", nil, []string{sh}, true, `al=/var/log/audit/audit.log; test -r "$al"`, nil, PassedWhenRetCodeFailed},
	{"VNFBPT50", true, "A non-root user must not sniff traffic with tcpdump", nil, []string{sh}, true, `(tcpdump -i lo -n 2>&1 & pid=$!;sleep 0.2;kill $pid)2>/dev/null | grep -i "listening on lo"`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT51", true, "A non-root user must not have access permissions to .htpasswd files", nil, []string{sh, find}, true, find_opts + `find / $find_opts -name "*.htpasswd" -print -exec cat {} +`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT52", true, "A non-root user must not have any ssh private keys stored in ssh-agent", nil, []string{sh, "ssh-add"}, true, `ssh-add -l | grep -iv "agent has no identities"`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT53", true, "A non-root user must not have access to gpg keys in gpg-agent", nil, []string{sh, "gpg-connect-agent"}, true, `gpg-connect-agent "keyinfo --list" /bye | grep "D - - 1"`, nil, PassedWhenRetCodeFailed},
	{"VNFBPT54", true, "A non-root user must not have access permissions to ssh-agent sockets of other users", nil, []string{sh, find, id}, true, `find /tmp/ -name 'agent.*' ! -user $(id -u) -readable -exec ls -l {} +`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT55", true, "A non-root user must not have access permissions to other users' gpg-agent sockets", nil, []string{sh, find}, true, `find /run/user/*/gnupg/ /tmp/ -type s -name 'S.gpg-agent*' ! -user $(id -u) -exec ls -l {} +`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT56", true, "A non-root user must not have access permissions to keepass database files of other users", nil, []string{sh, find}, true, find_opts + `find / $find_opts -regextype egrep -iregex ".*\.kdbx?" ! -user $(id -u) -readable -type f -print`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT57", true, "A non-root user must not have access permissions to .password-store directories of other users", nil, []string{sh, find}, true, find_opts + `find / $find_opts -name ".password-store" -type d ! -user $(id -u) -readable -exec ls -ld {} +`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT58", true, "A non-root user must not have access permissions to any Kerberos credentials files of other users", nil, []string{sh, find}, true, find_opts + `find / $find_opts -name "*.so" -prune -o \( -name "krb5cc*" -o -name "*.ccache" -o -name "*.kirbi" -o -name "*.keytab" \) -type f ! -user $(id -u) -readable -exec ls -lh {} +`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT66", true, "Check environment variables for secrets", []*Dependency{{Id: "ENUM06", Type: TestFunc, VarName: "env_variables"}}, []string{env}, false, ``, vnfbpt66, PassedWhenRetCodeSuccessful},

	{"VNFBPT67", true, "Cron tasks must not be owned by a non-root user", nil, []string{find, ls}, false, `find -L /etc/cron.allow /etc/cron.d /etc/cron.daily /etc/cron.deny /etc/cron.hourly /etc/cron.monthly /etc/crontab /etc/cron.weekly /etc/anacron /var/spool/cron -type f -not -uid 0 -exec ls -ld {} + 2>/dev/null`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT68", true, "Files and directories must not be world-writable", nil, []string{find}, false, `find / \( -path /sys -o -path /proc -o -path /dev \) -prune -o \( -type f -o -type d \) -perm -0002 ! -perm -1000 -exec ls -ld {} \; 2>/dev/null`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT69", true, "Log directories must not be writable by others", nil, []string{find}, false, `find /var/log -type d -perm -0002 -exec ls -ld {} \; 2>/dev/null`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT70", true, "Log files must not be accessible (readable) by others", nil, []string{find}, false, `find /var/log -type f -perm /0007 -exec ls -l {} \; 2>/dev/null`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT71", true, "Owner of a log file must be a system user", nil, []string{find}, false, `find /var/log -type f ! -uid 0 -exec ls -l {} \; 2>/dev/null`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT72", true, "Group of a log file must be a system group", nil, []string{find}, false, `find /var/log -type f ! \( -group 0 -o -group 4 \) -exec ls -l {} \; 2>/dev/null`, nil, PassedWhenStdoutEmpty},
	{"VNFBPT73", true, "It is forbidden to use known-exploitable binaries for privilege elevation with sudo", nil, []string{sudo}, false, `sudo -l`, nil, PassedWhenRetCodeSuccessful},
	//{"VNFBPT74", true, "A VM is forbidden to log secrets", []*Dependency{{Id: "ENUM06", Type: TestFunc, VarName: "env_variables"}}, []string{env}, false, ``, vnfbpt66, PassedWhenRetCodeSuccessful},
}

//var TestCasesDev = []*TestCase{
//	{"VNFBPT67", true, "User must not run known vulnerable sudo executables", []*Dependency{{Id: "ENUM07", Type: TestFunc, VarName: "sudo_output"}}, []string{sudo}, false, ``, vnfbpt67, PassedWhenRetCodeSuccessful},
//}

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
