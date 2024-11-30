package types

// Permission represents access control for documents
type Permission struct {
	Read  []string // List of roles that can read
	Write []string // List of roles that can write
}

// HasReadAccess checks if a role has read access
func (p *Permission) HasReadAccess(role string) bool {
	for _, r := range p.Read {
		if r == role {
			return true
		}
	}
	return false
}

// HasWriteAccess checks if a role has write access
func (p *Permission) HasWriteAccess(role string) bool {
	for _, r := range p.Write {
		if r == role {
			return true
		}
	}
	return false
}

// AddReadAccess adds a role to the read access list
func (p *Permission) AddReadAccess(role string) {
	if !p.HasReadAccess(role) {
		p.Read = append(p.Read, role)
	}
}

// AddWriteAccess adds a role to the write access list
func (p *Permission) AddWriteAccess(role string) {
	if !p.HasWriteAccess(role) {
		p.Write = append(p.Write, role)
	}
}

// RemoveReadAccess removes a role from the read access list
func (p *Permission) RemoveReadAccess(role string) {
	for i, r := range p.Read {
		if r == role {
			p.Read = append(p.Read[:i], p.Read[i+1:]...)
			return
		}
	}
}

// RemoveWriteAccess removes a role from the write access list
func (p *Permission) RemoveWriteAccess(role string) {
	for i, r := range p.Write {
		if r == role {
			p.Write = append(p.Write[:i], p.Write[i+1:]...)
			return
		}
	}
}
