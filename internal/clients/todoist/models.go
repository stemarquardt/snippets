package todoist

import "time"

type Project struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	CommentCount int    `json:"comment_count"`
	Order        int    `json:"order"`
	Color        string `json:"color"`
	Shared       bool   `json:"shared"`
	SyncID       int64  `json:"sync_id"`
	Favorite     bool   `json:"favorite"`
	InboxProject bool   `json:"inbox_project"`
	TeamInbox    bool   `json:"team_inbox"`
	URL          string `json:"url"`
}

// type Task struct {
// 	ID           string     `json:"id"`
// 	UserID       string     `json:"user_id"`
// 	ProjectID    string     `json:"project_id"`
// 	Content      string     `json:"content"`
// 	Description  string     `json:"description"`
// 	Due          *Due       `json:"due"`
// 	LabelIDs     []string   `json:"label_ids"`
// 	Order        int        `json:"order"`
// 	Priority     int        `json:"priority"`
// 	CommentCount int        `json:"comment_count"`
// 	CreatedAt    time.Time  `json:"created_at"`
// 	URL          string     `json:"url"`
// 	Completed    bool       `json:"completed"`
// 	CompletedAt  *time.Time `json:"completed_at,omitempty"`
// 	IsCollapsed  bool       `json:"is_collapsed"`
// }

type Task struct {
	ID             string                 `json:"id"`
	UserID         string                 `json:"user_id"`
	ProjectID      string                 `json:"project_id"`
	SectionID      string                 `json:"section_id,omitempty"`
	ParentID       string                 `json:"parent_id,omitempty"`
	AddedByUID     string                 `json:"added_by_uid,omitempty"`
	AssignedByUID  string                 `json:"assigned_by_uid,omitempty"`
	ResponsibleUID string                 `json:"responsible_uid,omitempty"`
	Labels         []string               `json:"labels,omitempty"`
	Deadline       map[string]interface{} `json:"deadline,omitempty"`
	Duration       map[string]int         `json:"duration,omitempty"`
	Checked        bool                   `json:"checked"`
	IsDeleted      bool                   `json:"is_deleted"`
	AddedAt        string                 `json:"added_at,omitempty"`
	CompletedAt    string                 `json:"completed_at,omitempty"`
	CompletedByUID string                 `json:"completed_by_uid,omitempty"`
	UpdatedAt      string                 `json:"updated_at,omitempty"`
	Due            *Due                   `json:"due,omitempty"`
	Priority       int                    `json:"priority"`
	ChildOrder     int                    `json:"child_order"`
	Content        string                 `json:"content"`
	Description    string                 `json:"description,omitempty"`
	NoteCount      int                    `json:"note_count"`
	DayOrder       int                    `json:"day_order"`
	IsCollapsed    bool                   `json:"is_collapsed"`
}

type Due struct {
	Date        string `json:"date"`
	Datetime    string `json:"datetime"`
	String      string `json:"string"`
	Timezone    string `json:"timezone"`
	IsRecurring bool   `json:"is_recurring"`
}

type CompletedInfo struct {
	Items []CompletedTask `json:"items"`
}

type CompletedTask struct {
	ID             string                 `json:"id"`
	UserID         string                 `json:"user_id"`
	ProjectID      string                 `json:"project_id"`
	SectionID      string                 `json:"section_id,omitempty"`
	ParentID       string                 `json:"parent_id,omitempty"`
	AddedByUID     string                 `json:"added_by_uid,omitempty"`
	AssignedByUID  string                 `json:"assigned_by_uid,omitempty"`
	ResponsibleUID string                 `json:"responsible_uid,omitempty"`
	Labels         []string               `json:"labels,omitempty"`
	Deadline       map[string]interface{} `json:"deadline,omitempty"`
	Duration       map[string]int         `json:"duration,omitempty"`
	Checked        bool                   `json:"checked"`
	IsDeleted      bool                   `json:"is_deleted"`
	AddedAt        string                 `json:"added_at,omitempty"`
	CompletedAt    string                 `json:"completed_at,omitempty"`
	CompletedByUID string                 `json:"completed_by_uid,omitempty"`
	UpdatedAt      string                 `json:"updated_at,omitempty"`
	Due            *Due                   `json:"due,omitempty"`
	Priority       int                    `json:"priority"`
	ChildOrder     int                    `json:"child_order"`
	Content        string                 `json:"content"`
	Description    string                 `json:"description,omitempty"`
	NoteCount      int                    `json:"note_count"`
	DayOrder       int                    `json:"day_order"`
	IsCollapsed    bool                   `json:"is_collapsed"`
}

type TodoistAPIOpts struct {
	ProjectID string    `json:"project_id,omitempty"`
	Since     time.Time `json:"since,omitempty"`
	Until     time.Time `json:"until,omitempty"`
	Limit     int       `json:"limit,omitempty"`
	Offset    int       `json:"offset,omitempty"`
}
