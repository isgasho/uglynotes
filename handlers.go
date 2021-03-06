package main

import (
	"encoding/json"
	"io/ioutil"
	"unicode/utf8"

	"github.com/ahui2016/uglynotes/model"
	"github.com/ahui2016/uglynotes/util"
	"github.com/gofiber/fiber/v2"
)

type (
	Note     = model.Note
	NoteType = model.NoteType
	History  = model.History
	TagGroup = model.TagGroup
)

func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	err = c.Status(code).JSON(fiber.Map{"message": err.Error()})
	if err != nil {
		// In case the c.JSON fails
		return c.Status(500).SendString("Internal Server Error")
	}
	return nil
}

func homePage(c *fiber.Ctx) error {
	return c.SendFile("./static/home.html")
}

func indexPage(c *fiber.Ctx) error {
	return c.SendFile("./static/index.html")
}

func loginPage(c *fiber.Ctx) error {
	return c.SendFile("./public/login.html")
}

func notePage(c *fiber.Ctx) error {
	return c.SendFile("./static/note.html")
}

func noteNewPage(c *fiber.Ctx) error {
	return c.SendFile("./static/note-edit.html")
}

func noteNewPage2(c *fiber.Ctx) error {
	return c.SendFile("./static/note-edit2.html")
}

func noteEditPage(c *fiber.Ctx) error {
	return c.SendFile("./static/note-edit.html")
}

func historyPage(c *fiber.Ctx) error {
	return c.SendFile("./static/history.html")
}

func historyPage2(c *fiber.Ctx) error {
	return c.SendFile("./static/history2.html")
}

func noteHistoryPage(c *fiber.Ctx) error {
	return c.SendFile("./static/note-history.html")
}

func tagPage(c *fiber.Ctx) error {
	return c.SendFile("./static/tag.html")
}

func tagsPage(c *fiber.Ctx) error {
	return c.SendFile("./static/tags.html")
}

func tagGroupsPage(c *fiber.Ctx) error {
	return c.SendFile("./static/tag-groups.html")
}

func converterPage(c *fiber.Ctx) error {
	return c.SendFile("./public/img-converter.html")
}

func searchPage(c *fiber.Ctx) error {
	return c.SendFile("./static/search.html")
}

func downloadDatabase(c *fiber.Ctx) error {
	return c.SendFile(dbPath)
}

func downloadDatabaseJSON(c *fiber.Ctx) error {
	return c.SendFile(exportPath)
}

func loginHandler(c *fiber.Ctx) error {
	if isLoggedIn(c) {
		return jsonMessage(c, "already logged in")
	}

	if c.FormValue("password") != config.Password {
		passwordTry++
		if err := checkPasswordTry(c); err != nil {
			return err
		}
		return jsonError(c, "Wrong Password", 400)
	}
	if err := db.Upgrade(); err != nil {
		return err
	}
	passwordTry = 0
	return db.SessionSet(c)
}

func checkLogin(c *fiber.Ctx) error {
	if isLoggedIn(c) {
		return jsonMessage(c, "OK")
	}
	return jsonMessage(c, "NG")
}

func getAllNotes(c *fiber.Ctx) error {
	notes, err := db2.AllNotes()
	if err != nil {
		return err
	}
	// trimContents(notes)
	return c.JSON(notes)
}

func getDeletedNotes(c *fiber.Ctx) error {
	notes, err := db2.AllDeletedNotes()
	if err != nil {
		return err
	}
	// trimContents(notes)
	return c.JSON(notes)
}

func exportAllNotes(c *fiber.Ctx) error {
	notes, err := db.AllNotesWithDeleted()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(exportPath, util.MustMarshalIndent(notes), 0600)
}

func trimContents(notes []Note) {
	for i := range notes {
		notes[i].Patches = nil
	}
}

func getNoteHandler(c *fiber.Ctx) error {
	note, err := db.GetByID(c.Params("id"))
	if err != nil {
		return err
	}
	return c.JSON(note)
}

func newNoteHandler(c *fiber.Ctx) error {
	db.Lock()
	defer db.Unlock()

	note, err := createNote(c)
	if err != nil {
		return jsonError(c, err.Error(), 400)
	}
	if err := db.Insert(note); err != nil {
		return err
	}
	return jsonMessage(c, note.ID)
}

func createNote(c *fiber.Ctx) (*Note, error) {
	noteType, err1 := getNoteType(c)
	title, err2 := getFormValue(c, "title")
	tags, err3 := getTags(c)
	patch := c.FormValue("patch") // 不能 TrimSpace!!
	if err := util.WrapErrors(err1, err2, err3); err != nil {
		return nil, err
	}
	note := db.NewNote(noteType)
	err1 = note.AddPatchSetTitle(patch, title)
	err2 = note.SetTags(tags)
	if err := util.WrapErrors(err1, err2); err != nil {
		return nil, err
	}
	return note, nil
}

func changeType(c *fiber.Ctx) error {
	db.Lock()
	defer db.Unlock()

	id := c.Params("id")
	noteType, err := getNoteType(c)
	if err != nil {
		return err
	}
	return db.ChangeType(id, noteType)
}

func updateNoteTags(c *fiber.Ctx) error {
	db.Lock()
	defer db.Unlock()

	id := c.Params("id")
	tags, err := getTags(c)
	if err != nil {
		return err
	}
	return db.UpdateTags(id, tags)
}

func patchNoteHandler(c *fiber.Ctx) error {
	db.Lock()
	defer db.Unlock()

	id := c.Params("id")
	patch := c.FormValue("patch") // 不能 TrimSpace!!
	title, err := getFormValue(c, "title")
	if err != nil {
		return err
	}

	count, err := db.AddPatchSetTitle(id, patch, title)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"message": count})
}

func notesSizeHandler(c *fiber.Ctx) error {
	size, err := db2.GetTotalSize()
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{
		"totalSize": size,
		"capacity":  config.DatabaseCapacity,
	})
}

func setTagGroupProtected(c *fiber.Ctx) error {
	db.Lock()
	defer db.Unlock()

	groupID := c.Params("id")
	protected, err := getProtected(c)
	if err != nil {
		return err
	}
	return db.SetTagGroupProtected(groupID, protected)
}

func shortHistories(histories []History) {
	for i := range histories {
		title := headLimit(histories[i].Contents, config.NoteTitleLimit)
		histories[i].Contents = title
	}
}

// headLimit 返回 s 开头限定长度的内容，其中 s 必须事先 TrimSpace 并确保不是空字串。
// 该函数会尽量确保最后一个字符是有效的 utf8 字符，但当限定长度的内容的全部字符都无效时，
// 则按原样返回限定长度的内容。
func headLimit(s string, limit int) (head string) {
	head = s
	if len(head) > limit {
		head = s[:limit]
	}
	for len(head) > 0 {
		if utf8.ValidString(head) {
			break
		}
		head = head[:len(head)-1]
	}
	if head == "" {
		head = s[:limit]
	}
	return head
}

func renameTag(c *fiber.Ctx) error {
	db.Lock()
	defer db.Unlock()

	oldName, err1 := getFormValue(c, "old-name")
	newName, err2 := getFormValue(c, "new-name")
	if err := util.WrapErrors(err1, err2); err != nil {
		return err
	}
	return db.RenameTag(oldName, newName)
}

func getNotesByTag(c *fiber.Ctx) error {
	tagName, err := getParams(c, "name")
	if err != nil {
		return err
	}
	notes, err := db.GetByTag(tagName)
	if err != nil {
		return err
	}
	return c.JSON(notes)
}

func getAllTags(c *fiber.Ctx) error {
	tags, err := db.AllTags()
	if err != nil {
		return err
	}
	return c.JSON(tags)
}

func allTagsByDate(c *fiber.Ctx) error {
	tags, err := db.AllTagsByDate()
	if err != nil {
		return err
	}
	return c.JSON(tags)
}

func allTagGroups(c *fiber.Ctx) error {
	groups, err := db2.AllTagGroups()
	if err != nil {
		return err
	}
	return c.JSON(groups)
}

// TODO: 如果只有一个标签，则不使用 db.SearchTagGroup
func searchTagGroup(c *fiber.Ctx) error {
	tags, err := getTagGroup(c)
	if err != nil {
		return err
	}
	notes, err := db.SearchTagGroup(tags)
	if err != nil {
		return err
	}
	trimContents(notes)
	return c.JSON(notes)
}

func searchTitle(c *fiber.Ctx) error {
	pattern, err := getParams(c, "pattern")
	if err != nil {
		return err
	}
	notes, err := db.SearchTitle(pattern)
	if err != nil {
		return err
	}
	trimContents(notes)
	return c.JSON(notes)
}

func addTagGroup(c *fiber.Ctx) error {
	db.Lock()
	defer db.Unlock()

	tags, err := getTags(c)
	if err != nil {
		return err
	}
	group := model.NewTagGroup(tags)
	if err := db.SaveTagGroup(group); err != nil {
		return err
	}
	return c.JSON(group)
}

func deleteTagGroup(c *fiber.Ctx) error {
	db.Lock()
	defer db.Unlock()

	groupID := c.Params("id")
	return db.DB.DeleteStruct(&TagGroup{ID: groupID})
}

func setNoteDeleted(c *fiber.Ctx) error {
	db.Lock()
	defer db.Unlock()

	id := c.Params("id")
	deleted, err := getDeleted(c)
	if err != nil {
		return err
	}
	return db.SetNoteDeleted(id, deleted)
}

func deleteNoteForever(c *fiber.Ctx) error {
	db.Lock()
	defer db.Unlock()

	id := c.Params("id")
	return db.DeleteNoteForever(id)
}

func deleteTag(c *fiber.Ctx) error {
	db.Lock()
	defer db.Unlock()

	name, err := getParams(c, "name")
	if err != nil {
		return err
	}
	return db.DeleteTag(name)
}

func deleteNoteHistories(c *fiber.Ctx) error {
	db.Lock()
	defer db.Unlock()

	id := c.Params("id")
	return db.DeleteNoteHistory(id)
}

func importNotes(c *fiber.Ctx) error {
	var notes []Note
	blob, err := ioutil.ReadFile(exportPath)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(blob, &notes); err != nil {
		return err
	}
	return db2.ImportNotes(notes)
}
