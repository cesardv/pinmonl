package model

type Sharetag struct {
	ID          string       `json:"id"`
	ShareID     string       `json:"shareId"`
	TagID       string       `json:"tagId"`
	Kind        SharetagKind `json:"kind"`
	ParentID    string       `json:"parentId"`
	Level       int          `json:"level"`
	Status      Status       `json:"status"`
	HasChildren bool         `json:"hasChildren"`

	Share    *Share        `json:"share,omitempty"`
	Tag      *Tag          `json:"tag,omitempty"`
	Children *SharetagList `json:"children,omitempty"`
}

type SharetagKind int

const (
	SharetagMust SharetagKind = iota
	SharetagAny
)

func IsValidSharetagKind(k SharetagKind) bool {
	checks := map[SharetagKind]int{
		SharetagMust: 0,
		SharetagAny:  0,
	}
	_, ok := checks[k]
	return ok
}

type SharetagList []*Sharetag

func (sl SharetagList) Keys() []string {
	keys := make([]string, 0)
	for _, s := range sl {
		keys = append(keys, s.ID)
	}
	return keys
}

func (sl SharetagList) Tags() TagList {
	tags := make([]*Tag, len(sl))
	for i := range sl {
		tags[i] = sl[i].Tag
	}
	return tags
}

func (sl SharetagList) ViewTags() TagList {
	tags := make([]*Tag, len(sl))
	for i := range sl {
		st := sl[i]
		tag := st.Tag
		tag.ParentID = st.ParentID
		tag.Level = st.Level
		tag.HasChildren = st.HasChildren
		tags[i] = tag
	}
	return tags
}
