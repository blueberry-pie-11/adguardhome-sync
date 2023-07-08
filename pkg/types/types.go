package types

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bakito/adguardhome-sync/pkg/versions"
	"go.uber.org/zap"
)

const (
	// DefaultAPIPath default api path
	DefaultAPIPath = "/control"
)

// Config application configuration struct
// +k8s:deepcopy-gen=true
type Config struct {
	Origin     AdGuardInstance   `json:"origin" yaml:"origin"`
	Replica    *AdGuardInstance  `json:"replica,omitempty" yaml:"replica,omitempty"`
	Replicas   []AdGuardInstance `json:"replicas,omitempty" yaml:"replicas,omitempty"`
	Cron       string            `json:"cron,omitempty" yaml:"cron,omitempty"`
	RunOnStart bool              `json:"runOnStart,omitempty" yaml:"runOnStart,omitempty"`
	API        API               `json:"api,omitempty" yaml:"api,omitempty"`
	Features   Features          `json:"features,omitempty" yaml:"features,omitempty"`
}

// API configuration
type API struct {
	Port     int    `json:"port,omitempty" yaml:"port,omitempty"`
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
	DarkMode bool   `json:"darkMode,omitempty" yaml:"darkMode,omitempty"`
}

// UniqueReplicas get unique replication instances
func (cfg *Config) UniqueReplicas() []AdGuardInstance {
	dedup := make(map[string]AdGuardInstance)
	if cfg.Replica != nil && cfg.Replica.URL != "" {
		dedup[cfg.Replica.Key()] = *cfg.Replica
	}
	for _, replica := range cfg.Replicas {
		if replica.URL != "" {
			dedup[replica.Key()] = replica
		}
	}

	var r []AdGuardInstance
	for _, replica := range dedup {
		if replica.APIPath == "" {
			replica.APIPath = DefaultAPIPath
		}
		r = append(r, replica)
	}
	return r
}

// Log the current config
func (cfg *Config) Log(l *zap.SugaredLogger) {
	c := cfg.DeepCopy()
	c.Origin.Mask()
	if c.Replica != nil {
		if c.Replica.URL == "" {
			c.Replica = nil
		} else {
			c.Replica.Mask()
		}
	}
	for i := range c.Replicas {
		c.Replicas[i].Mask()
	}
	l.With("config", c).Debug("Using config")
}

// AdGuardInstance AdguardHome config instance
// +k8s:deepcopy-gen=true
type AdGuardInstance struct {
	URL                string `json:"url" yaml:"url"`
	APIPath            string `json:"apiPath,omitempty" yaml:"apiPath,omitempty"`
	Username           string `json:"username,omitempty" yaml:"username,omitempty"`
	Password           string `json:"password,omitempty" yaml:"password,omitempty"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify" yaml:"insecureSkipVerify"`
	AutoSetup          bool   `json:"autoSetup" yaml:"autoSetup"`
	InterfaceName      string `json:"interfaceName" yaml:"interfaceName"`
	DHCPServerEnabled  *bool  `json:"dhcpServerEnabled,omitempty" yaml:"dhcpServerEnabled,omitempty"`
}

// Key AdGuardInstance key
func (i *AdGuardInstance) Key() string {
	return fmt.Sprintf("%s#%s", i.URL, i.APIPath)
}

// Mask maks username and password
func (i *AdGuardInstance) Mask() {
	i.Username = mask(i.Username)
	i.Password = mask(i.Password)
}

func mask(s string) string {
	if s == "" {
		return "***"
	}
	return fmt.Sprintf("%v***%v", string(s[0]), string(s[len(s)-1]))
}

// Protection API struct
type Protection struct {
	ProtectionEnabled bool `json:"protection_enabled"`
}

// Status API struct
type Status struct {
	Protection
	DNSAddresses  []string `json:"dns_addresses"`
	DNSPort       int      `json:"dns_port"`
	HTTPPort      int      `json:"http_port"`
	DhcpAvailable bool     `json:"dhcp_available"`
	Running       bool     `json:"running"`
	Version       string   `json:"version"`
	Language      string   `json:"language"`
}

// Filters list of Filter
type Filters []Filter

// Merge merge Filters
func (f Filters) Merge(other Filters) (Filters, Filters, Filters) {
	current := make(map[string]Filter)

	var adds Filters
	var updates Filters
	var removes Filters
	for _, f := range f {
		current[f.URL] = f
	}

	for i := range other {
		rr := other[i]
		if c, ok := current[rr.URL]; ok {
			if !c.Equals(&rr) {
				updates = append(updates, rr)
			}
			delete(current, rr.URL)
		} else {
			adds = append(adds, rr)
		}
	}

	for _, rr := range current {
		removes = append(removes, rr)
	}

	return adds, updates, removes
}

// Filter API struct
type Filter struct {
	ID         int    `json:"id"`
	Enabled    bool   `json:"enabled"`
	URL        string `json:"url"`  // needed for add
	Name       string `json:"name"` // needed for add
	RulesCount int    `json:"rules_count"`
	Whitelist  bool   `json:"whitelist"` // needed for add
}

// Equals Filter equal check
func (f *Filter) Equals(o *Filter) bool {
	return f.Enabled == o.Enabled && f.URL == o.URL && f.Name == o.Name
}

// FilterUpdate  API struct
type FilterUpdate struct {
	URL       string `json:"url"`
	Data      Filter `json:"data"`
	Whitelist bool   `json:"whitelist"`
}

// FilteringStatus API struct
type FilteringStatus struct {
	FilteringConfig
	Filters          Filters   `json:"filters"`
	WhitelistFilters Filters   `json:"whitelist_filters"`
	UserRules        UserRules `json:"user_rules"`
}

// UserRules API struct
type UserRules []string

// String toString of Users
func (ur UserRules) String() string {
	return strings.Join(ur, "\n")
}

// ToPayload return the version specific payload for user rules
func (ur UserRules) ToPayload(version string) interface{} {
	if versions.IsNewerThan(version, versions.LastStringCustomRules) {
		return &UserRulesRequest{Rules: ur}
	}
	return ur.String()
}

// UserRulesRequest API struct
type UserRulesRequest struct {
	Rules UserRules
}

// String toString of Users
func (ur UserRulesRequest) String() string {
	return ur.Rules.String()
}

// EnableConfig API struct
type EnableConfig struct {
	Enabled bool `json:"enabled"`
}

// IntervalConfig API struct
type IntervalConfig struct {
	Interval float64 `json:"interval"`
}

// FilteringConfig API struct
type FilteringConfig struct {
	EnableConfig
	IntervalConfig
}

// QueryLogConfig API struct
type QueryLogConfig struct {
	EnableConfig
	IntervalConfig
	AnonymizeClientIP bool `json:"anonymize_client_ip"`
}

// Equals QueryLogConfig equal check
func (qlc *QueryLogConfig) Equals(o *QueryLogConfig) bool {
	return qlc.Enabled == o.Enabled && qlc.AnonymizeClientIP == o.AnonymizeClientIP && qlc.Interval == o.Interval
}

// RefreshFilter API struct
type RefreshFilter struct {
	Whitelist bool `json:"whitelist"`
}

// Services API struct
type Services []string

// Sort sort Services
func (s Services) Sort() {
	sort.Strings(s)
}

// Equals Services equal check
func (s Services) Equals(o Services) bool {
	s.Sort()
	o.Sort()
	return equals(s, o)
}

func equals(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// InstallConfig AdguardHome install config
type InstallConfig struct {
	Web      InstallPort `json:"web"`
	DNS      InstallPort `json:"dns"`
	Username string      `json:"username"`
	Password string      `json:"password"`
}

// InstallPort AdguardHome install config port
type InstallPort struct {
	IP         string `json:"ip"`
	Port       int    `json:"port"`
	Status     string `json:"status"`
	CanAutofix bool   `json:"can_autofix"`
}
