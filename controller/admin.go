package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JohanAanesen/NTNU-Bachelor-Management-System-For-CS-Assignments/model"
	"github.com/JohanAanesen/NTNU-Bachelor-Management-System-For-CS-Assignments/shared/db"
	"github.com/JohanAanesen/NTNU-Bachelor-Management-System-For-CS-Assignments/shared/session"
	"github.com/JohanAanesen/NTNU-Bachelor-Management-System-For-CS-Assignments/shared/view"
	_ "github.com/go-sql-driver/mysql" //database driver
	"github.com/rs/xid"
	"github.com/shurcooL/github_flavored_markdown"
	"html/template"
	"io"
	"log"
	"net/http"
	"time"
)

// AdminGET handles GET-request at /admin
func AdminGET(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	v := view.New(r)
	v.Name = "admin/index"

	assignmentRepo := model.AssignmentRepository{}
	assignments, err := assignmentRepo.GetAllToUserSorted(session.GetUserFromSession(r).ID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		log.Println(err)
		return
	}

	courses, err := model.GetCoursesToUser(session.GetUserFromSession(r).ID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		log.Println(err)
		return
	}

	v.Vars["Courses"] = courses
	v.Vars["Assignments"] = assignments

	v.Render(w)
}

// AdminCourseGET handles GET-request at /admin/course
func AdminCourseGET(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	v := view.New(r)
	v.Name = "admin/course/index"

	courses, err := model.GetCoursesToUser(session.GetUserFromSession(r).ID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		log.Println(err)
		return
	}

	v.Vars["Courses"] = courses

	v.Render(w)
}

// AdminCreateCourseGET handles GET-request at /admin/course/create
func AdminCreateCourseGET(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	v := view.New(r)
	v.Name = "admin/course/create"

	v.Render(w)
}

// AdminCreateCoursePOST handles POST-request at /admin/course/create
// Inserts a new course to the database
func AdminCreateCoursePOST(w http.ResponseWriter, r *http.Request) {
	//check if user is already logged in
	user := session.GetUserFromSession(r)

	course := model.Course{
		Hash:        xid.NewWithTime(time.Now()).String(),
		Code:        r.FormValue("code"),
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		Year:        r.FormValue("year"),
		Semester:    r.FormValue("semester"),
	}

	// TODO (Svein): Move this to model, in a function
	//insert into database
	result, err := db.GetDB().Exec("INSERT INTO course(hash, coursecode, coursename, year, semester, description, teacher) VALUES(?, ?, ?, ?, ?, ?, ?)",
		course.Hash, course.Code, course.Name, course.Year, course.Semester, course.Description, user.ID)

	// Log error
	if err != nil {
		//todo log error
		log.Println(err.Error())
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	// Get course id
	id, err := result.LastInsertId()
	if err != nil {
		log.Println(err.Error())
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	// Convert from int64 to int
	course.ID = int(id)

	// Log createCourse in the database and give error if something went wrong
	lodData := model.Log{UserID: user.ID, Activity: model.CreatedCourse, CourseID: course.ID}
	if !model.LogToDB(lodData) {
		log.Fatal("Could not save createCourse log to database! (admin.go)")
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	// Add user to course
	if !model.AddUserToCourse(user.ID, course.ID) {
		log.Println("Could not add user to course! (admin.go)")
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	// Log joinedCourse in the db and give error if something went wrong
	lodData = model.Log{UserID: user.ID, Activity: model.JoinedCourse, CourseID: course.ID}
	if !model.LogToDB(lodData) {
		log.Fatal("Could not save createCourse log to database! (admin.go)")
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	//IndexGET(w, r) //success redirect to homepage
	http.Redirect(w, r, "/", http.StatusFound) //success redirect to homepage
}

// AdminUpdateCourseGET handles GET-request at /admin/course/update/{id}
func AdminUpdateCourseGET(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	v := view.New(r)
	v.Name = "admin/course/update"

	v.Render(w)
}

// AdminUpdateCoursePOST handles POST-request at /admin/course/update/{id}
func AdminUpdateCoursePOST(w http.ResponseWriter, r *http.Request) {

}

// AdminSubmissionGET handles GET-request to /admin/submission
func AdminSubmissionGET(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	var subRepo = model.SubmissionRepository{}

	submissions, err := subRepo.GetAll()
	if err != nil {
		log.Println(err)
		return
	}

	v := view.New(r)
	v.Name = "admin/submission/index"

	v.Vars["Submissions"] = submissions

	v.Render(w)
}

// AdminSubmissionCreateGET handles GET-request to /admin/submission/create
func AdminSubmissionCreateGET(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	v := view.New(r)
	v.Name = "admin/submission/create"

	v.Render(w)
}

// AdminSubmissionCreatePOST handles POST-request to /admin/submission/create
func AdminSubmissionCreatePOST(w http.ResponseWriter, r *http.Request) {
	// Get data from the form
	data := r.FormValue("data")
	// Declare Form-struct
	var form = model.Form{}
	// Unmarshal the JSON-string sent from the form
	err := json.Unmarshal([]byte(data), &form)
	if err != nil {
		log.Println(err)
		return
	}
	// Declare empty slice for error messages
	var errorMessages []string

	// Check form name
	if form.Name == "" {
		errorMessages = append(errorMessages, "Form name cannot be blank.")
	}

	// Check number of fields
	if len(form.Fields) == 0 {
		errorMessages = append(errorMessages, "Form needs to have at least 1 field.")
	}

	// Check if any error messages has been appended
	if len(errorMessages) != 0 {
		// TODO (Svein): Keep data from the previous submit
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		v := view.New(r)
		v.Name = "admin/submission/create"

		v.Vars["Errors"] = errorMessages

		v.Render(w)

		return
	}

	// Declare an empty Repository for Submission
	var repo = model.SubmissionRepository{}
	// Insert data to database
	err = repo.Insert(form)
	if err != nil {
		log.Println(err)
		return
	}

	// Redirect to /admin/submission
	http.Redirect(w, r, "/admin/submission", http.StatusFound)
}

// DatetimeLocalToRFC3339 converts a string from datetime-local HTML input-field to time.Time object
func DatetimeLocalToRFC3339(str string) (time.Time, error) {
	// TODO (Svein): Move this to a utils.go or something
	if str == "" {
		return time.Time{}, errors.New("error: could not parse empty datetime-string")
	}
	if len(str) < 16 {
		return time.Time{}, errors.New("cannot convert a string less then 16 characters: DatetimeLocalToRFC3339()")
	}
	year := str[0:4]
	month := str[5:7]
	day := str[8:10]
	hour := str[11:13]
	min := str[14:16]

	value := fmt.Sprintf("%s-%s-%sT%s:%s:00Z", year, month, day, hour, min)
	return time.Parse(time.RFC3339, value)
}

// AdminFaqGET handles GET-request at admin/faq/index
func AdminFaqGET(w http.ResponseWriter, r *http.Request) {
	content := model.GetDateAndQuestionsFAQ()
	if content.Questions == "-1" {
		log.Println("Something went wrong with getting the faq (admin.go)")
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	// Convert to html
	questions := github_flavored_markdown.Markdown([]byte(content.Questions))

	v := view.New(r)
	v.Name = "admin/faq/index"
	v.Vars["Updated"] = content.Date.Format("02. January 2006 - 15:04")
	v.Vars["Questions"] = template.HTML(questions)

	v.Render(w)
}

// AdminFaqEditGET returns the edit view for the faq
func AdminFaqEditGET(w http.ResponseWriter, r *http.Request) {
	//
	content := model.GetDateAndQuestionsFAQ()
	if content.Questions == "-1" {
		log.Println("Something went wrong with getting the faq (admin.go)")
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	v := view.New(r)
	v.Name = "admin/faq/edit"
	v.Vars["Updated"] = content.Date.Format("02. January 2006 - 15:04")
	v.Vars["RawContent"] = content.Questions

	v.Render(w)
}

// AdminFaqUpdatePOST handles the edited markdown faq
func AdminFaqUpdatePOST(w http.ResponseWriter, r *http.Request) {
	// Check that the questions arrived
	updatedFAQ := r.FormValue("rawQuestions")
	if updatedFAQ == "" {
		log.Println("Form is empty! (admin.go)")
		ErrorHandler(w, r, http.StatusBadRequest)
		return
	}

	// Check that it's possible to get the old faq from db
	content := model.GetDateAndQuestionsFAQ()
	if content.Questions == "-1" {
		log.Println("Something went wrong with getting the faq (admin.go)")
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	// Check that it's changes to the new faq
	if content.Questions == updatedFAQ {
		log.Println("Old and new faq can not be equal! (admin.go)")
		ErrorHandler(w, r, http.StatusBadRequest)
		return
	}

	// Check that it went okay to add new faq to db
	if !model.UpdateFAQ(updatedFAQ) {
		log.Println("Something went wrong with updating the faq! (admin.go)")
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	// Get user for logging purposes
	user := session.GetUserFromSession(r)

	// Collect the log data
	logData := model.Log{
		UserID:   user.ID,
		Activity: model.UpdateAdminFAQ,
		OldValue: content.Questions,
		NewValue: updatedFAQ,
	}

	// Log that a teacher has changed the faq
	if !model.LogToDB(logData) {
		log.Println("Something went wrong with logging the new faq! (admin.go)")
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	AdminFaqGET(w, r)
}

// AdminSettingsGET handles GET-request at admin/setting
func AdminSettingsGET(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	v := view.New(r)
	v.Name = "admin/settings/index"

	v.Render(w)
}

// AdminSettingsPOST handles POST-request at admin/setting
func AdminSettingsPOST(w http.ResponseWriter, r *http.Request) {
	// TODO (Svein): Handle incoming data
	http.Redirect(w, r, "/admin/settings", http.StatusOK)
}

// AdminSettingsImportGET handles GET-request at admin/setting/import
func AdminSettingsImportGET(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	v := view.New(r)
	v.Name = "admin/settings/import"

	v.Render(w)
}

// AdminSettingsImportPOST handles POST-request at admin/setting/import
func AdminSettingsImportPOST(w http.ResponseWriter, r *http.Request) {
	var buffer bytes.Buffer
	r.ParseMultipartForm(32 << 20)
	file, _, err := r.FormFile("db_import")
	if err != nil {
		log.Println(err)
		return
	}

	defer file.Close()
	defer buffer.Reset()

	_, err = io.Copy(&buffer, file)
	if err != nil {
		log.Println(err)
		return
	}

	content := buffer.String()
	fmt.Println(content)

	// TODO (Svein): Backup of current DB
	// TODO (Svein): Query this file
	// TODO (Svein): Save the world

	http.Redirect(w, r, "/admin/settings", http.StatusOK)
}

// GoToHTMLDatetimeLocal converts time.Time object to 'datetime-local' in HTML
func GoToHTMLDatetimeLocal(t time.Time) string {
	day := fmt.Sprintf("%d", t.Day())
	month := fmt.Sprintf("%d", t.Month())
	year := fmt.Sprintf("%d", t.Year())
	hour := fmt.Sprintf("%d", t.Hour())
	minute := fmt.Sprintf("%d", t.Minute())

	if t.Day() < 10 {
		day = "0" + day
	}

	if t.Month() < 10 {
		month = "0" + month
	}

	if t.Hour() < 10 {
		hour = "0" + hour
	}

	if t.Minute() < 10 {
		minute = "0" + minute
	}

	return fmt.Sprintf("%s-%s-%sT%s:%s", year, month, day, hour, minute)
}
