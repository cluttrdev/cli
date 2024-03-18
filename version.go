package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

type VersionInfo interface {
	// The version number
	Version() string

	// The revision identifier of the commit
	Revision() string

	// The modification time of the commit, in RFC3339 format
	Time() string

	// Whether the source tree had uncommitted local changes
	Modified() bool

	// The version of the Go toolchain that built the binary
	GoVersion() string
}

type BuildInfo struct {
	buildInfo *debug.BuildInfo
	version   string
}

func NewBuildInfo(version string) *BuildInfo {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		info = &debug.BuildInfo{}
	}

	return &BuildInfo{
		buildInfo: info,
		version:   version,
	}
}

func (bi *BuildInfo) Version() string {
	if bi.version != "" {
		return bi.version
    } else if v := bi.pseudoVersion(); v != "" {
        return v
    }

	return bi.buildInfo.Main.Version
}

func (bi *BuildInfo) Revision() string {
	for _, setting := range bi.buildInfo.Settings {
		if setting.Key == "vcs.revision" {
			return setting.Value
		}
	}
	return ""
}

func (bi *BuildInfo) Time() string {
	for _, setting := range bi.buildInfo.Settings {
		if setting.Key == "vcs.time" {
			return setting.Value
		}
	}
	return ""
}

func (bi *BuildInfo) Modified() bool {
	for _, setting := range bi.buildInfo.Settings {
		if setting.Key == "vcs.modified" {
			v, err := strconv.ParseBool(setting.Value)
			if err != nil {
				return false
			}
			return v
		}
	}
	return false
}

func (bi *BuildInfo) GoVersion() string {
	return bi.buildInfo.GoVersion
}

func (bi *BuildInfo) pseudoVersion() string {
    t, err := time.Parse(time.RFC3339, bi.Time())
    if err != nil {
        return ""
    }
    timestamp := t.Format("060102030405")
    revision := bi.Revision()[:12]
    return fmt.Sprintf("v0.0.0-%s-%s", timestamp, revision)
}

func DefaultVersionInfo() VersionInfo {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		info = &debug.BuildInfo{}
	}

	return &BuildInfo{
		buildInfo: info,
	}
}

func NewVersionCommand(info VersionInfo, out io.Writer) *Command {
	cfg := versionCmdConfig{
		version: info,
		flags:   flag.NewFlagSet("version", flag.ExitOnError),
		out:     out,
	}

	if cfg.out == nil {
		cfg.out = os.Stdout
	}

	cfg.RegisterFlags(cfg.flags)

	return &Command{
		Name:      "version",
		ShortHelp: "Show version information",
		Flags:     cfg.flags,
		Exec:      cfg.Exec,
	}
}

func DefaultVersionCommand(out io.Writer) *Command {
	info := DefaultVersionInfo()
	return NewVersionCommand(info, out)
}

type versionCmdConfig struct {
	version VersionInfo

	flags *flag.FlagSet

	out io.Writer
}

func (c *versionCmdConfig) RegisterFlags(fs *flag.FlagSet) {
	a := fs.Bool("all", false, "print all information")
	fs.BoolVar(a, "a", false, "shorthand option for `--all`")

	n := fs.Bool("number", false, "print the version number")
	fs.BoolVar(n, "n", false, "shorthand option for `--number`")
	r := fs.Bool("revision", false, "print the commit revision identifier")
	fs.BoolVar(r, "r", false, "shorthand option for `--revision`")
	t := fs.Bool("time", false, "print the commit revision modification time")
	fs.BoolVar(t, "t", false, "shorthand option for `--time`")
	m := fs.Bool("modified", false, "print the commit revision identifier")
	fs.BoolVar(m, "m", false, "shorthand option for `--modified`")
	g := fs.Bool("go-version", false, "print the Go toolchain version")
	fs.BoolVar(g, "g", false, "shorthand option for `--go-version`")

	fs.Bool("json", false, "print information in JSON")
}

func (c *versionCmdConfig) Exec(ctx context.Context, args []string) error {
	any := false
	c.flags.Visit(func(f *flag.Flag) {
		if f.Name != "json" {
			any = true
		}
	})
	all := testFlag(c.flags, "all")

	if testFlag(c.flags, "json") {
		return c.writeJson(any, all)
	}
	return c.writeText(any, all)
}

func testFlag(fs *flag.FlagSet, name string) bool {
	f := fs.Lookup(name)
	if f == nil {
		return false
	}

	v, err := strconv.ParseBool(f.Value.String())
	if err != nil {
		return false
	}

	return v
}

func (c *versionCmdConfig) writeText(any bool, all bool) error {
	builder := strings.Builder{}

	if !any || testFlag(c.flags, "number") || all {
		builder.WriteString(c.version.Version())
	}
	if testFlag(c.flags, "revision") || all {
		builder.WriteString(fmt.Sprintf(" %s", c.version.Revision()))
	}
	if testFlag(c.flags, "time") || all {
		builder.WriteString(fmt.Sprintf(" %s", c.version.Time()))
	}
	if testFlag(c.flags, "go-version") || all {
		builder.WriteString(fmt.Sprintf(" %s", c.version.GoVersion()))
	}
	if testFlag(c.flags, "modified") || all {
		if c.version.Modified() {
			builder.WriteString(" (modified)")
		}
	}

	s := builder.String()

	_, err := fmt.Fprintln(c.out, strings.TrimSpace(s))
	if err != nil {
		return fmt.Errorf("error writing version information: %w", err)
	}
	return nil
}

func (c *versionCmdConfig) writeJson(any bool, all bool) error {
	data := map[string]string{}

	if !any || testFlag(c.flags, "number") || all {
		data["Version"] = c.version.Version()
	}
	if testFlag(c.flags, "revision") || all {
		data["Revision"] = c.version.Revision()
	}
	if testFlag(c.flags, "time") || all {
		data["Time"] = c.version.Time()
	}
	if testFlag(c.flags, "go-version") || all {
		data["GoVersion"] = c.version.GoVersion()
	}
	if testFlag(c.flags, "modified") || all {
		data["Modified"] = fmt.Sprint(c.version.Modified())
	}

	m, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error encoding version information: %w", err)
	}

	_, err = fmt.Fprintln(c.out, string(m))
	if err != nil {
		return fmt.Errorf("error writing version information: %w", err)
	}
	return nil
}
