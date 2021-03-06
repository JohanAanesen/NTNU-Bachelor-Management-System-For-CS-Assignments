package controller

import (
	"github.com/JohanAanesen/CSAMS/webservice/model"
	"github.com/JohanAanesen/CSAMS/webservice/service"
	"github.com/JohanAanesen/CSAMS/webservice/shared/db"
	"github.com/JohanAanesen/CSAMS/webservice/shared/session"
	"github.com/JohanAanesen/CSAMS/webservice/shared/util"
	"github.com/JohanAanesen/CSAMS/webservice/shared/view"
	"log"
	"net/http"
	"regexp"
	"time"
)

// IndexGET serves homepage to authenticated users, send anonymous to login
func IndexGET(w http.ResponseWriter, r *http.Request) {
	// Current User
	currentUser := session.GetUserFromSession(r)

	// Services
	services := service.NewServices(db.GetDB())

	//get courses to user
	courses, err := services.Course.FetchAllForUserOrdered(currentUser.ID)
	if err != nil {
		log.Println(err)
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	// Need custom struct to get the coursecode and delivery status
	type ActiveAssignment struct {
		Assignment model.Assignment
		CourseCode string
		Delivered  bool
		Reviews    int
	}

	var activeAssignments []ActiveAssignment
	var noOfReviewsLeft int

	for _, course := range courses { //iterate all courses
		assignments, err := services.Assignment.FetchFromCourse(course.ID) //get assignments from course
		if err != nil {
			log.Println("services assignment, fetch from course", err)
			ErrorHandler(w, r, http.StatusInternalServerError)
			return
		}

		// TODO time-norwegian
		timeNow := util.GetTimeInCorrectTimeZone()
		for _, assignment := range assignments { //go through all it's assignments again
			if timeNow.After(assignment.Publish) && timeNow.Before(assignment.ReviewDeadline.Add(time.Hour*12)) { // Assignments stay on front page until an half day after review deadline is over

				// Initiate variable
				delivered := false

				// Only check if the user isn't a teacher
				if !currentUser.Teacher {
					// Check if student has submitted assignment
					delivered, err = services.SubmissionAnswer.HasUserSubmitted(assignment.ID, currentUser.ID)
					if err != nil {
						log.Println("services, submission answer, has user submitted", err.Error())
						ErrorHandler(w, r, http.StatusInternalServerError)
						return
					}

					// Check if user is in the peer review table
					inReviewTable, err := services.PeerReview.TargetExists(assignment.ID, currentUser.ID)
					if err != nil {
						log.Println("services, peer review, target exists", err.Error())
						ErrorHandler(w, r, http.StatusInternalServerError)
						return
					}

					// If its -404 the user doesn't exists in the peer review table
					noOfReviewsLeft = -404

					// Only check for count if user exists in th peer review table
					if inReviewTable {
						// Get number of reviews done bu user
						reviewDone, err := services.ReviewAnswer.CountReviewsDone(currentUser.ID, assignment.ID)
						if err != nil {
							log.Println("services, review answer, countreviews reviewDone", err.Error())
							ErrorHandler(w, r, http.StatusInternalServerError)
							return
						}

						// Calculate how many left
						noOfReviewsLeft = int(assignment.Reviewers.Int64) - reviewDone
					}
				}
				activeAssignments = append(activeAssignments, ActiveAssignment{Assignment: *assignment, CourseCode: course.Code, Delivered: delivered, Reviews: noOfReviewsLeft})
			}
		}
	}

	// Set values
	v := view.New(r)
	v.Name = "index"

	v.Vars["isStudent"] = !currentUser.Teacher
	v.Vars["Courses"] = courses
	v.Vars["Assignments"] = activeAssignments
	v.Vars["Message"] = session.GetFlash(w, r)

	v.Render(w)
}

// JoinCoursePOST adds user to course
func JoinCoursePOST(w http.ResponseWriter, r *http.Request) {
	// Get user
	currentUser := session.GetUserFromSession(r)
	// Services
	services := service.NewServices(db.GetDB())

	hash := r.FormValue("courseID")

	rgex := regexp.MustCompile("[a-zA-Z0-9]{20}")
	result := rgex.FindAllString(hash, -1)

	courseExist := true
	var course *model.Course

	// go through and check for hash in string + save it
	for _, element := range result {
		course = services.Course.Exists(element)
		if course.ID == -1 {
			courseExist = false
		} else {
			hash = element
		}
	}

	// Give feedback if course does not exist
	if !courseExist {
		log.Println("course does not exist")
		ErrorHandler(w, r, http.StatusBadRequest)
		return
	}

	existsInCourse, err := services.Course.UserInCourse(currentUser.ID, course.ID)
	if err != nil {
		log.Println("services, course, user in course", err.Error())
		ErrorHandler(w, r, http.StatusBadRequest)
		return
	}

	// Check that user isn't in this class
	if existsInCourse {
		log.Println("user already in course!")
		ErrorHandler(w, r, http.StatusBadRequest)
		return
	}

	err = services.Course.AddUser(currentUser.ID, course.ID)
	if err == service.ErrUserAlreadyInCourse {
		log.Println("user already in course", service.ErrUserAlreadyInCourse)
		http.Redirect(w, r, "/", http.StatusFound) //success, redirect to homepage
		return
	}

	// Add user to course if possible
	if err != nil {
		log.Println("error when adding user to course", err.Error())
		ErrorHandler(w, r, http.StatusBadRequest)
		return
	}

	// Log user join course
	err = services.Logs.InsertJoinCourse(currentUser.ID, course.ID)
	if err != nil {
		log.Println("log, join course, ", err.Error())
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	// Give feedback to user
	_ = session.SetFlash("You joined "+course.Code+" - "+course.Name, w, r)

	//IndexGET(w, r)
	http.Redirect(w, r, "/", http.StatusFound) //success redirect to homepage
}
