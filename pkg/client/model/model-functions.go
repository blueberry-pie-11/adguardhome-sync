package model

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/bakito/adguardhome-sync/pkg/utils"
	"github.com/jinzhu/copier"
)

// Clone the config
func (c *DhcpStatus) Clone() *DhcpStatus {
	clone := &DhcpStatus{}
	_ = copier.Copy(clone, c)
	return clone
}

// Equals dhcp server config equal check
func (c *DhcpStatus) Equals(o *DhcpStatus) bool {
	a, _ := json.Marshal(c)
	b, _ := json.Marshal(o)
	return string(a) == string(b)
}

func (c *DhcpStatus) HasConfig() bool {
	return (c.V4 != nil && c.V4.isValid()) || (c.V6 != nil && c.V6.isValid())
}

func (j DhcpConfigV4) isValid() bool {
	return j.GatewayIp != nil && j.SubnetMask != nil && j.RangeStart != nil && j.RangeEnd != nil
}

func (j DhcpConfigV6) isValid() bool {
	return j.RangeStart != nil
}

type DhcpStaticLeases []DhcpStaticLease

// MergeDhcpStaticLeases the leases
func MergeDhcpStaticLeases(l *[]DhcpStaticLease, other *[]DhcpStaticLease) (DhcpStaticLeases, DhcpStaticLeases) {
	var thisLeases []DhcpStaticLease
	var otherLeases []DhcpStaticLease

	if l != nil {
		thisLeases = *l
	}
	if other != nil {
		otherLeases = *other
	}
	current := make(map[string]DhcpStaticLease)

	var adds DhcpStaticLeases
	var removes DhcpStaticLeases
	for _, le := range thisLeases {
		current[le.Mac] = le
	}

	for _, le := range otherLeases {
		if _, ok := current[le.Mac]; ok {
			delete(current, le.Mac)
		} else {
			adds = append(adds, le)
		}
	}

	for _, rr := range current {
		removes = append(removes, rr)
	}

	return adds, removes
}

// Equals dns config equal check
func (c *DNSConfig) Equals(o *DNSConfig) bool {
	cc := c.Clone()
	oo := o.Clone()
	cc.Sort()
	oo.Sort()

	a, _ := json.Marshal(cc)
	b, _ := json.Marshal(oo)
	return string(a) == string(b)
}

func (c *DNSConfig) Clone() *DNSConfig {
	return utils.Clone(c, &DNSConfig{})
}

// Sort sort dns config
func (c *DNSConfig) Sort() {
	if c.UpstreamDns != nil {
		sort.Strings(*c.UpstreamDns)
	}

	if c.UpstreamDns != nil {
		sort.Strings(*c.BootstrapDns)
	}

	if c.UpstreamDns != nil {
		sort.Strings(*c.LocalPtrUpstreams)
	}
}

// Equals access list equal check
func (al *AccessList) Equals(o *AccessList) bool {
	return equals(al.AllowedClients, o.AllowedClients) &&
		equals(al.DisallowedClients, o.DisallowedClients) &&
		equals(al.BlockedHosts, o.BlockedHosts)
}

func equals(a *[]string, b *[]string) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	aa := *a
	bb := *b
	if len(aa) != len(bb) {
		return false
	}
	for i, v := range aa {
		if v != bb[i] {
			return false
		}
	}
	return true
}

// Sort clients
func (cl *Client) Sort() {
	if cl.Ids != nil {
		sort.Strings(*cl.Ids)
	}
	if cl.Tags != nil {
		sort.Strings(*cl.Tags)
	}
	if cl.BlockedServices != nil {
		sort.Strings(*cl.BlockedServices)
	}
	if cl.Upstreams != nil {
		sort.Strings(*cl.Upstreams)
	}
}

// Equals Clients equal check
func (cl *Client) Equals(o *Client) bool {
	cl.Sort()
	o.Sort()

	a, _ := json.Marshal(cl)
	b, _ := json.Marshal(o)
	return string(a) == string(b)
}

// Add ac client
func (clients *Clients) Add(cl Client) {
	if clients.Clients == nil {
		clients.Clients = &ClientsArray{cl}
	} else {
		a := append(*clients.Clients, cl)
		clients.Clients = &a
	}
}

// Merge merge Clients
func (clients *Clients) Merge(other *Clients) ([]*Client, []*Client, []*Client) {
	current := make(map[string]*Client)
	if clients.Clients != nil {
		cc := *clients.Clients
		for i := range cc {
			client := cc[i]
			current[*client.Name] = &client
		}
	}

	expected := make(map[string]*Client)
	if other.Clients != nil {
		oc := *other.Clients
		for i := range oc {
			client := oc[i]
			expected[*client.Name] = &client
		}
	}

	var adds []*Client
	var removes []*Client
	var updates []*Client

	for _, cl := range expected {
		if oc, ok := current[*cl.Name]; ok {
			if !cl.Equals(oc) {
				updates = append(updates, cl)
			}
			delete(current, *cl.Name)
		} else {
			adds = append(adds, cl)
		}
	}

	for _, rr := range current {
		removes = append(removes, rr)
	}

	return adds, updates, removes
}

// Key RewriteEntry key
func (re *RewriteEntry) Key() string {
	var d string
	var a string
	if re.Domain != nil {
		d = *re.Domain
	}
	if re.Answer != nil {
		a = *re.Answer
	}
	return fmt.Sprintf("%s#%s", d, a)
}

// RewriteEntries list of RewriteEntry
type RewriteEntries []RewriteEntry

// Merge RewriteEntries
func (rwe *RewriteEntries) Merge(other *RewriteEntries) (RewriteEntries, RewriteEntries, RewriteEntries) {
	current := make(map[string]RewriteEntry)

	var adds RewriteEntries
	var removes RewriteEntries
	var duplicates RewriteEntries
	processed := make(map[string]bool)
	for _, rr := range *rwe {
		if _, ok := processed[rr.Key()]; !ok {
			current[rr.Key()] = rr
			processed[rr.Key()] = true
		} else {
			// remove duplicate
			removes = append(removes, rr)
		}
	}

	for _, rr := range *other {
		if _, ok := current[rr.Key()]; ok {
			delete(current, rr.Key())
		} else {
			if _, ok := processed[rr.Key()]; !ok {
				adds = append(adds, rr)
				processed[rr.Key()] = true
			} else {
				//	skip duplicate
				duplicates = append(duplicates, rr)
			}
		}
	}

	for _, rr := range current {
		removes = append(removes, rr)
	}

	return adds, removes, duplicates
}

func MergeFilters(this *[]Filter, other *[]Filter) ([]Filter, []Filter, []Filter) {
	if this == nil && other == nil {
		return nil, nil, nil
	}

	current := make(map[string]*Filter)

	var adds []Filter
	var updates []Filter
	var removes []Filter
	if this != nil {
		for i := range *this {
			fi := (*this)[i]
			current[fi.Url] = &fi
		}
	}

	if other != nil {
		for i := range *other {
			rr := (*other)[i]
			if c, ok := current[rr.Url]; ok {
				if !c.Equals(&rr) {
					updates = append(updates, rr)
				}
				delete(current, rr.Url)
			} else {
				adds = append(adds, rr)
			}
		}
	}

	for _, rr := range current {
		removes = append(removes, *rr)
	}

	return adds, updates, removes
}

// Equals Filter equal check
func (f *Filter) Equals(o *Filter) bool {
	return f.Enabled == o.Enabled && f.Url == o.Url && f.Name == o.Name
}