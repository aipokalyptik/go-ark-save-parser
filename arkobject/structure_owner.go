package arkobject

func (s Structure) IsOwnedBy(owner ObjectOwner) bool {
	if s.Owner.PlayerID != 0 && s.Owner.PlayerID == owner.PlayerID {
		return true
	}
	if s.Owner.PlayerName != "" && s.Owner.PlayerName == owner.PlayerName {
		return true
	}
	if s.Owner.TribeName != "" && s.Owner.TribeName == owner.TribeName {
		return true
	}
	if s.Owner.TribeID != 0 && s.Owner.TribeID == owner.TribeID {
		return true
	}
	if s.Owner.OriginalPlacerID != 0 && s.Owner.OriginalPlacerID == owner.OriginalPlacerID {
		return true
	}
	return false
}
