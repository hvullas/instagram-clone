package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"os"
	"path/filepath"

	"net/http"

	_ "image/jpeg"
	_ "image/png"

	_ "github.com/lib/pq"
	"github.com/robfig/cron/v3"
	"golang.org/x/crypto/bcrypt"
)

// db connection info
const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "Ullas123"
	dbname   = "instagram_db"
)

// struct for new user registration
type NewUser struct {
	UserName       string  `json:"user_name"`
	Password       string  `json:"password"`
	Name           string  `json:"name"`
	Email          string  `json:"email"`
	PhoneNumber    string  `json:"phone_number"`
	DOB            string  `json:"DOB"`
	Private        *bool   `json:"private_account"`
	Bio            *string `json:"bio"`
	DisplayPicture string  `json:"display_picture"`
}

type UserName struct {
	UserName string `json:"user_name"`
}

type GetProfilePicURL struct {
	PicURL string `json:"picURL"`
}

// struct to insert a post to db
type InsertPost struct {
	UserID          *int64   `json:"user_id"`
	PostPath        []string `json:"post_path"`
	PostCaption     *string  `json:"post_caption"`
	HashtagIds      []int64  `json:"hashtag_ids"`
	TaggedIds       []int64  `json:"tagged_ids"`
	Location        *string  `json:"location"`
	HideLikeCount   *bool    `json:"hide_like_count"`
	TurnOffComments *bool    `json:"turnoff_comments"`
}

// returned postid by postMedia api
type PostId struct {
	PostId int64 `json:"post_id"`
}
type UserID struct {
	UserId int64 `json:"user_id"`
}

// struct to get user posts and post by postId
type UsersPost struct {
	UserID            int64    `json:"user_iD"`
	UserName          string   `json:"user_name"`
	UserProfilePicURL string   `json:"user_profile_picUrl"`
	PostId            int64    `json:"post_id"`
	PostURL           []string `json:"postURL"`
	FileType          string   `json:"file_type"`
	AttachedLocation  string   `json:"attached_location"`
	LikeStatus        bool     `json:"like_status"`
	Likes             int64    `json:"likes"`
	PostCaption       string   `json:"caption"`
	HideLikeCount     bool     `json:"hide_like_count"`
	TurnOffComments   bool     `json:"turnoff_comments"`
	SavedStatus       bool     `json:"saved_post"`
	PostedOn          string   `json:"posted_on"`
}

// posting like to a post
type LikePost struct {
	PostId int64 `json:"post_id"`
	UserID int64 `json:"user_id"`
}

// to get count of likes on a post
type TotalLikes struct {
	TotalLikes int64 `json:"total_likes"`
}

// to post a comment
type CommentBody struct {
	PostId      int64  `json:"post_id"`
	UserID      int64  `json:"user_id"`
	CommentBody string `json:"comment_body"`
}

// comment id returned after succefull comment insertion
type ReturnedCommentId struct {
	ReturnedCommentId int64 `json:"returned_commentId"`
}

// to get the comments by post id
type CommentsOfPost struct {
	CommentId           int64  `json:"comment_id"`
	CommentorUserName   string `json:"commentor_user_name"`
	CommentorDisplayPic string `json:"commentor_display_pic"`
	PostId              int64  `json:"post_id"`
	CommentBody         string `json:"comment_body"`
	CommentedOn         string `json:"commented_on"`
}

// to delete a comment
type DeleteComment struct {
	UserID    int64 `json:"user_id"`
	PostId    int64 `json:"post_id"`
	CommentId int64 `json:"comment_id"`
}

// login
type LoginCred struct {
	UserName string `josn:"user_name"`
	Password string `josn:"password"`
}

// to follow another user
type Follow struct {
	MyId      int64 `json:"my_id"`
	Following int64 `json:"following_id"`
}

type FollowStatus struct {
	FollowStatus bool `json:"follow_status"`
}
type SavedStatus struct {
	SavedStatus bool `json:"saved_status"`
}

// to update bio of profile
type ProfileUpdate struct {
	UserID   int64  `json:"user_id"`
	Name     string `json:"name"`
	UserName string `json:"user_name"`
	Bio      string `json:"bio"`
}

// to give profile info response
type Profile struct {
	UserID         int64  `json:"user_id"`
	UserName       string `json:"user_name"`
	PrivateAccount bool   `json:"private_account"`
	PostCount      int64  `json:"post_count"`
	FollowerCount  int64  `json:"follower_count"`
	FollowingCount int64  `json:"following_count"`
	Bio            string `json:"bio"`
	ProfilePic     string `json:"profile_picURL"`
}

// to get all follower and following
type Follows struct {
	UserID              int64  `json:"user_id"`
	UserName            string `json:"user_name"`
	Name                string `json:"name"`
	ProfilePic          string `json:"profile_pic"`
	FollowingBackStatus bool   `json:"following_back_status"`
}

// to serve search results
type Accounts struct {
	UserID     int64  `json:"user_id"`
	UserName   string `json:"user_name"`
	Name       string `json:"name"`
	ProfilePic string `json:"profile_pic"`
}

// saved posts response
type SavedPosts struct {
	PostId      int64  `json:"post_id"`
	PostURL     string `json:"postURL"`
	ContentType string `json:"content_type"`
}

// to serve follow requests
type FollowRequest struct {
	UserID     int64  `json:"requestor_user_id"`
	UserName   string `json:"requestor_user_name"`
	ProfilePic string `json:"requestor_profile_pic"`
	CreatedOn  string `json:"request_created_on"`
	Accepted   bool   `json:"accepted"`
}
type FollowAcceptance struct {
	AcceptorUserID int64 `json:"acceptor_user_id"`
	RequestorId    int64 `json:"requestor_user_id"`
	AcceptStatus   bool  `json:"acceptance_status"`
}
type DeleteFollower struct {
	MyuserId       int64 `json:"my_user_id"`
	FollowerUserId int64 `json:"follower_user_id"`
}

type HashtagSearch struct {
	Hashtag string `json:"hashtag"`
}

type HashtagSearchResult struct {
	HashId    int64  `json:"hash_id"`
	HashName  string `json:"hash_name"`
	PostCount int64  `json:"total_posts"`
}
type Newhashtag struct {
	NewHashId int64 `json:"new_hash_tag_id"`
}
type StoryInfo struct {
	UserID    int64     `json:"user_id"`
	TaggedIds [][]int64 `json:"tagged_ids"`
}
type StoryMedia struct {
	StoryId int64 `json:"story_id"`
}

type ReturnedStoryId struct {
	ReturnedStoryId int64 `json:"returned_story_id"`
	PostAsStory     bool  `json:"post_as_story"`
}

type UploadStory struct {
	StoryId  int64 `json:"story_id"`
	Uploaded bool  `json:"uploaded"`
}

type GetStory struct {
	StoryId   int64   `json:"story_id"`
	StoryURL  string  `json:"storyurl"`
	PostedOn  string  `json:"posted_on"`
	Success   bool    `json:"upload_status"`
	TaggedIds []int64 `json:"tagged_userids"`
	FileType  string  `json:"file_type"`
}
type PostAsStory struct {
	UserID    int64   `json:"user_id"`
	TaggedIds []int64 `json:"tagged_ids"`
	StoryURL  string  `json:"storyURL"`
}

type ActiveStories struct {
	User_id        int64   `json:"user_id"`
	User_name      string  `json:"user_name"`
	Profile_picURL string  `json:"profile_pic_url"`
	Story_id       []int64 `json:"story_ids"`
	Seen_status    bool    `json:"story_seen_status"`
}

type UpdateStorySeenStatus struct {
	UserID  int64 `json:"user_id"`
	StoryId int64 `json:"story_id"`
}

// func to get the file extensions(used while serving files)
func GetExtension(extension string) string {
	switch extension {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".bmp":
		return "image/bmp"
	case ".mp4":
		return "video/mp4"
	case ".mov":
		return "video/quicktime"
	default:
		return ""
	}
}

const MB = 1 << 20

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	//create profilePhoto directory if not exists
	if err := os.MkdirAll("./profilePhoto", os.ModePerm); err != nil {
		log.Fatal("Error creating posts directory", err)
	}

	//create posts directory if not exists
	if err := os.MkdirAll("./posts", os.ModePerm); err != nil {
		log.Fatal("Error creating posts directory", err)
	}

	//create stories directory if not exists
	if err := os.MkdirAll("./stories", os.ModePerm); err != nil {
		log.Fatal("Error creating posts directory", err)
	}
	//cron delete stories after 24 hours

	//cron delete posts which are not updated with media
	cron := cron.New()

	cron.AddFunc("29 16 * * *", func() {

		//_,err=db.Query("DELETE FROM posts WHERE complete_post=$1",false)
		_, err = db.Query("DELETE FROM example WHERE timestamp<= current_timestamp - interval '10 minute'")
		if err != nil {
			log.Panic(err)
			return
		}
		log.Println("cron active")
	})
	cron.Start()

	http.HandleFunc("/newUserInfo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var userdata NewUser
		err = json.NewDecoder(r.Body).Decode(&userdata)
		if err != nil {
			fmt.Fprintln(w, "Error decoding Request body")
			return
		}

		//check for missing fields
		if userdata.Private == nil {
			http.Error(w, "Invalid or missing privateAccount field", http.StatusMethodNotAllowed)
			return
		}
		if userdata.UserName == "" {
			http.Error(w, "Missing userName field", http.StatusMethodNotAllowed)
			return
		}
		if userdata.Password == "" {
			http.Error(w, "Missing password field", http.StatusMethodNotAllowed)
			return
		}
		if userdata.Email == "" {
			http.Error(w, "Missing email field", http.StatusMethodNotAllowed)
			return
		}
		if userdata.PhoneNumber == "" {
			http.Error(w, "Missing phone number field", http.StatusMethodNotAllowed)
			return
		}
		if userdata.UserName == "" {
			http.Error(w, "Missing userName field", http.StatusMethodNotAllowed)
			return
		}
		if userdata.DOB == "" {
			http.Error(w, "Missing DOB field", http.StatusMethodNotAllowed)
			return
		}
		if userdata.Bio == nil {
			http.Error(w, "Missing bio field", http.StatusMethodNotAllowed)
			return
		}
		if userdata.Name == "" {
			http.Error(w, "Missing name field", http.StatusMethodNotAllowed)
			return
		}

		// user name validation

		match, _ := regexp.MatchString("^[a-zA-Z0-9][a-zA-Z0-9_]*$", userdata.UserName)
		if !match {
			fmt.Fprintln(w, "User name should start with alphabet and can have combination minimum 8 characters of numbers and only underscore(_)")
			return
		}

		if len(userdata.UserName) < 7 || len(userdata.UserName) > 20 {
			http.Error(w, "Username should be of length(7,20)", http.StatusMethodNotAllowed)
			return
		}

		if len(userdata.Name) > 20 {
			http.Error(w, "Name should be less than 20 character", http.StatusMethodNotAllowed)
			return
		}

		// user password validation
		if len(userdata.Password) == 0 {
			http.Error(w, "Missing password field", http.StatusMethodNotAllowed)
			return
		}

		match, _ = regexp.MatchString("[0-9]+?", userdata.Password)
		if !match {
			fmt.Fprintln(w, "Password must contain atleast one number")
			return
		}
		match, _ = regexp.MatchString("[A-Z]+?", userdata.Password)
		if !match {
			fmt.Fprintln(w, "Password must contain atleast upper case letter")
			return
		}
		match, _ = regexp.MatchString("[a-z]+?", userdata.Password)
		if !match {
			fmt.Fprintln(w, "Password must contain atleast lower case letter")
			return
		}
		match, _ = regexp.MatchString("[!@#$%^&*_]+?", userdata.Password)
		if !match {
			fmt.Fprintln(w, "Password must contain atleast special character")
			return
		}
		match, _ = regexp.MatchString(".{8,30}", userdata.Password)
		if !match {
			fmt.Fprintln(w, "Password length must be atleast 8 character long")
			return
		}

		//phone number validation
		match, _ = regexp.MatchString("^[+]{1}[0-9]{0,3}\\s?[0-9]{10}$", userdata.PhoneNumber)
		if !match {
			fmt.Fprintln(w, "Please enter valid phone number")
			return
		}

	

		//validate email using net/mail
		emailregex := regexp.MustCompile("^[A-Za-za0-9.!#$%&'*+\\/=?^_`{|}~-]+@[A-Za-z](?:[A-Za-z0-9-]{0,61}[A-Za-z])?(?:\\.[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?)*$")
		match = emailregex.MatchString(userdata.Email)
		if !match {
			fmt.Fprintln(w, "Please enter valid email")
			return
		}
		if len(userdata.Email) < 3 && len(userdata.Email) > 254 {
			http.Error(w, "Invalid email", http.StatusMethodNotAllowed)
			return
		}

		i := strings.Index(userdata.Email, "@")
		host := userdata.Email[i+1:]

		_, err := net.LookupMX(host)
		if err != nil {
			http.Error(w, "Invalid email(host not found)", http.StatusMethodNotAllowed)
			return
		}
		//validate date
		layout := "2006-01-02"
		bdate, err := time.Parse(layout, userdata.DOB)
		if err != nil {
			fmt.Fprintln(w, "Enter a valid date format")
			return
		}
		cdate := time.Now()

		age := cdate.Sub(bdate)
		if age.Hours() < 113958 {
			fmt.Fprintln(w, "Enter proper date of birth,You ahould be minimum of 13 years old to create an account")
			return
		}

		//check for duplication of user name
		userExists := `SELECT user_name FROM users WHERE user_name=$1`
		var usernameexits string
		err = db.QueryRow(userExists, userdata.UserName).Scan(&usernameexits)
		// if err != nil {
		// 	panic(err)
		// }
		if usernameexits == userdata.UserName {
			fmt.Fprintln(w, "User Name already exists. Try another user name")
			return
		}

		//check for duplication of email address
		userEmailExists := `SELECT email FROM users WHERE email=$1`
		var emailexits string
		err = db.QueryRow(userEmailExists, userdata.Email).Scan(&emailexits)
		// if err != nil {
		// 	panic(err)
		// }
		if emailexits == userdata.Email {
			fmt.Fprintln(w, "Account with this email already exists")
			return
		}

		//check for duplication of phone number
		userCellExists := `SELECT phone_number FROM users WHERE phone_number=$1`
		var numberExists string
		err = db.QueryRow(userCellExists, userdata.PhoneNumber).Scan(&numberExists)
		// if err != nil {
		// 	panic(err)
		// }
		if numberExists == userdata.PhoneNumber {
			fmt.Fprintln(w, "Account with this phone number already exists")
			return
		}

		//hashing the password before storing to the database
		pass := []byte(userdata.Password)

		// Hashing the password
		hash, err := bcrypt.GenerateFromPassword(pass, 8)
		if err != nil {
			panic(err)
		}

		userdata.DisplayPicture = "profilePhoto/DefaultProfilePicture.jpeg"

		regUserInfo := `INSERT INTO users (user_name,password,email,phone_number,dob,bio,private,display_pic,name) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING user_id`
		var userID UserID
		err = db.QueryRow(regUserInfo, userdata.UserName, string(hash), userdata.Email, userdata.PhoneNumber, userdata.DOB, userdata.Bio, userdata.Private, userdata.DisplayPicture, userdata.Name).Scan(&userID.UserId)
		if err != nil {
			panic(err)

		}

		json.NewEncoder(w).Encode(userID)

	})

	//user authorisation
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var login LoginCred
		err := json.NewDecoder(r.Body).Decode(&login)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusMethodNotAllowed)
			return
		}
		//auth
		var passwordHash string
		err = db.QueryRow("SELECT password FROM users WHERE user_name=$1", login.UserName).Scan(&passwordHash)
		if err != nil {
			panic(err)
		}
		err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(login.Password))
		if err == nil {
			fmt.Fprintln(w, true)
		}
		if err != nil {
			fmt.Fprintln(w, "Invalid password")
			return
		}

	})

	//handle function to upload users' display picture //ullas
	http.HandleFunc("/updateUserDisplayPic", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// set parse data limit
		r.Body = http.MaxBytesReader(w, r.Body, 5*MB)
		err := r.ParseMultipartForm(5 * MB) // 10 MB
		if err != nil {
			http.Error(w, "Unable to parse form", http.StatusMethodNotAllowed)
			return
		}

		// Get the file from the request
		file, fileHeader, err := r.FormFile("display_picture")
		if err != nil {
			http.Error(w, "Missing formfile", http.StatusBadRequest)
			return
		}

		//get cleaned file name
		s := regexp.MustCompile(`\s+`).ReplaceAllString(fileHeader.Filename, "")
		time := fmt.Sprintf("%v", time.Now())
		s = regexp.MustCompile(`\s+`).ReplaceAllString(time, "") + s

		jsonData := r.FormValue("user_id")
		var userId UserID
		err = json.Unmarshal([]byte(jsonData), &userId)
		if err != nil {
			http.Error(w, "Error unmarshalling JSON data", http.StatusInternalServerError)
			return
		}
		var deleteUrl string
		db.QueryRow("SELECT display_pic FROM users WHERE user_id=$1", userId.UserId).Scan(&deleteUrl)
		filelocation := "./" + deleteUrl

		if filelocation != "./profilePhoto/DefaultProfilePicture.jpeg" {
			os.Remove(filelocation)
		}

		//check for file allowed file format
		match, _ := regexp.MatchString("^.*\\.(jpg|JPG|png|PNG|JPEG|jpeg|bmp|BMP)$", s)
		if !match {
			fmt.Fprintln(w, "Only JPG,JPEG,PNG,BMP formats are allowed for upload")
			return
		} else {
			//check for the file size
			if size := fileHeader.Size; size > 8*MB {
				http.Error(w, "File size exceeds 8MB", http.StatusInternalServerError)
				return
			}
		}

		// Create a new file on the server(folder)
		fileName := s

		dst, err := os.Create(filepath.Join("./profilePhoto", fileName))
		if err != nil {
			http.Error(w, "Unable to create file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		// Copy the file data to the directory
		_, err = io.Copy(dst, file)
		if err != nil {
			http.Error(w, "Unable to write file", http.StatusInternalServerError)
			return
		}

		filePath := filepath.Join("./profilePhoto", fileName)

		// image, _, err := image.DecodeConfig(file)
		// if err != nil {
		// 	panic(err)
		// }

		// if image.Height < 150 && image.Width < 150 {
		// 	http.Error(w, "Image resolution too low", http.StatusInternalServerError)
		// 	e := os.Remove(filePath)
		// 	if e != nil {
		// 		panic(e)
		// 	}

		// 	return
		// }

		urlpart1 := "http://localhost:3000/getProfilePic/"

		var retrivedUrl string

		var dpURL GetProfilePicURL
		err = db.QueryRow("UPDATE users SET display_pic=$1 WHERE user_id=$2 RETURNING display_pic", filePath, userId.UserId).Scan(&retrivedUrl)
		if err != nil {
			panic(err)
		}

		dpURL.PicURL = urlpart1 + retrivedUrl
		err = json.NewEncoder(w).Encode(dpURL)
		if err != nil {
			panic(err)
		}

	})

	//func to get profilePic
	http.HandleFunc("/getProfilePic/profilePhoto/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		url := fmt.Sprint(r.URL)

		_, file := path.Split(url)

		imagePath := "./profilePhoto/" + file
		imagedata, err := ioutil.ReadFile(imagePath)

		if err != nil {
			http.Error(w, "Couldn't read the file", http.StatusInternalServerError)
			return
		}

		ext := strings.ToLower(filepath.Ext(file))

		contentType := GetExtension(ext)

		if contentType == "" {
			http.Error(w, "Unsupported file format", http.StatusUnsupportedMediaType)
			return
		}

		w.Header().Set("Content-Type", contentType)

		_, err = w.Write(imagedata)
		if err != nil {
			http.Error(w, "failed to write image data to response", http.StatusInternalServerError)
			return
		}

	})

	//handle func to serve posts
	http.HandleFunc("/download/posts/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		url := fmt.Sprint(r.URL)

		_, file := path.Split(url)

		imagePath := "./posts/" + file

		imagedata, err := ioutil.ReadFile(imagePath)
		if err != nil {
			http.Error(w, "Couldn't read the file", http.StatusInternalServerError)
			return
		}

		ext := strings.ToLower(filepath.Ext(file))

		contentType := GetExtension(ext)

		if contentType == "" {
			http.Error(w, "Unsupported file format", http.StatusUnsupportedMediaType)
			return
		}

		w.Header().Set("Content-Type", contentType)

		_, err = w.Write(imagedata)
		if err != nil {
			http.Error(w, "failed to write image data to response", http.StatusInternalServerError)
			return
		}
	})

	//to post media to instagram
	http.HandleFunc("/postMediaInfo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var postInfo InsertPost
		err = json.NewDecoder(r.Body).Decode(&postInfo)
		if err != nil {
			fmt.Fprintln(w, "Check input data field formats")
			return
		}

		//check for missing fields
		if postInfo.TurnOffComments == nil || postInfo.HideLikeCount == nil || postInfo.Location == nil || postInfo.UserID == nil || postInfo.PostCaption == nil {
			http.Error(w, "Missing field/fields in the request", http.StatusMethodNotAllowed)
			return
		}

		//validate input user id

		match, _ := regexp.MatchString("^.*[0-9]$", strconv.Itoa(int(*postInfo.UserID)))
		if !match {
			fmt.Fprintln(w, "check input post id format")
			return
		}

		if len(postInfo.HashtagIds) > 30 {
			http.Error(w, "You can use only 30 hashtags in the caption", http.StatusInternalServerError)
			return
		}
		if len(postInfo.TaggedIds) > 20 {
			http.Error(w, "Only 20 users can be tagged", http.StatusInternalServerError)
			return
		}

		var idexists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE user_id=$1)", postInfo.UserID).Scan(&idexists)
		if err != nil {
			http.Error(w, "Invalid user-id", http.StatusInternalServerError)
			return
		}

		if !idexists {
			http.Error(w, "No user exists with this user-id", http.StatusInternalServerError)
			return
		}

		//check for existance of tagged ids

		for _, id := range postInfo.TaggedIds {
			err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE user_id=$1)", id).Scan(&idexists)
			if err != nil {
				http.Error(w, "Invalid user-id", http.StatusBadRequest)
				return
			}
			if !idexists {
				http.Error(w, "No user exists with this tagged id", http.StatusBadRequest)
				fmt.Fprint(w, id)
				return
			}
		}

		//check for existance of hash ids
		for _, id := range postInfo.HashtagIds {
			err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM hashtags WHERE hash_id=$1)", id).Scan(&idexists)
			if err != nil {
				http.Error(w, "Invalid hash-id", http.StatusInternalServerError)
				return
			}
			if !idexists {
				http.Error(w, "Invalid hash-id", http.StatusInternalServerError)
				return
			}
		}

		// //validate input location format
		if *postInfo.Location != "" {
			pointRegex := regexp.MustCompile(`^[-+]?([1-8]?\d(\.\d+)?|90(\.0+)?),\s*[-+]?(180(\.0+)?|((1[0-7]\d)|([1-9]?\d))(\.\d+)?)$`)
			if !pointRegex.MatchString(*postInfo.Location) {
				fmt.Fprintln(w, "check input location format")
				return
			}
		} else {
			*postInfo.Location = "0,0" //stuff user location when there is access
		}

		//check for max length of post caption 150 chars
		if len(*postInfo.PostCaption) > 2200 {
			http.Error(w, "Max allowed length of post caption is 2200 character", http.StatusNotAcceptable)
			return
		}

		var postId PostId
		insertPostInfo := `INSERT INTO posts(user_id,poat_caption,location,hide_like,hide_comments) VALUES($1,$2,$3,$4,$5) RETURNING post_id`
		err = db.QueryRow(insertPostInfo, postInfo.UserID, postInfo.PostCaption, postInfo.Location, postInfo.HideLikeCount, postInfo.TurnOffComments).Scan(&postId.PostId)
		if err != nil {
			panic(err)
		}

		//update tags
		for _, tagid := range postInfo.TaggedIds {
			_, err = db.Query("INSERT INTO tagged_users(post_id,tagged_ids) VALUES($1,$2)", postId.PostId, tagid)
			if err != nil {
				db.Query("DELETE FROM posts WHERE post_id=$1", postId.PostId)
				http.Error(w, "Error inserting tagged users", http.StatusInternalServerError)
				return
			}

		}

		//check for hashtags

		for _, hashid := range postInfo.HashtagIds {
			_, err = db.Query("INSERT INTO mentions(hash_id,post_id) VALUES($1,$2)", hashid, postId.PostId)
			if err != nil {
				fmt.Println(err)
				db.Query("DELETE FROM tagged_users WHERE post_id=$1", postId.PostId)
				http.Error(w, "Error inserting mentions", http.StatusInternalServerError)
				return
			}
		}

		json.NewEncoder(w).Encode(postId)

		fmt.Fprintf(w, "Posts uploaded successfully.")

	})

	//handle func to upload user posts
	http.HandleFunc("/postMediaPath", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 4096*MB)
		err := r.ParseMultipartForm(4096 * MB)
		if err != nil {
			http.Error(w, "Error parsing multipart form data", http.StatusInternalServerError)
			return
		}
		jsonData := r.FormValue("postId")

		var postId PostId

		err = json.Unmarshal([]byte(jsonData), &postId)
		if err != nil {

			http.Error(w, "Error unmarshalling JSON data,Enter correct postID", http.StatusInternalServerError)
			return
		}

		var tempid int64
		err = db.QueryRow("SELECT post_id FROM posts WHERE post_id=$1", postId.PostId).Scan(&tempid)
		if err != nil {
			http.Error(w, "Invalid post_id", http.StatusInternalServerError)
			return
		}

		if tempid != postId.PostId {
			http.Error(w, "Invalid post_id", http.StatusInternalServerError)
			return
		}

		var postPath []string
		fileHeaders := r.MultipartForm.File
		if len(fileHeaders) == 0 {
			http.Error(w, "No files attached", http.StatusInternalServerError)
			return
		}

		for _, fileHeaders := range fileHeaders {
			for _, fileHeader := range fileHeaders {
				if len(fileHeaders) > 10 {
					http.Error(w, "Only 10 files allowed", http.StatusInternalServerError)
					return
				}
				file, err := fileHeader.Open()
				if err != nil {
					http.Error(w, "Unable to open the file", http.StatusInternalServerError)
					return
				}
				defer file.Close()

				//check for file allowed file format
				match, _ := regexp.MatchString("^.*\\.(jpg|JPG|png|PNG|JPEG|jpeg|bmp|BMP|MP4|mp4|mov|MOV|GIF|gif)$", fileHeader.Filename)
				if !match {
					fmt.Fprintln(w, "Only JPG,JPEG,PNG,BMP formats are allowed for upload")
					return
				} else {
					//check for the file size
					if size := fileHeader.Size; size > 8*MB {
						http.Error(w, "File size exceeds 8MB", http.StatusInternalServerError)
						return
					}
				}

				// image, _, err := image.DecodeConfig(file)
				// if err != nil {
				// 	http.Error(w, "Cannot read the image configs", http.StatusInternalServerError)
				// 	return
				// }

				// if image.Height < 155 && image.Width < 155 {
				// 	http.Error(w, "Image resolution too low", http.StatusInternalServerError)
				// 	return
				// }

				// fmt.Fprintln(w, fileHeader.Filename, ":", image.Width, "x", image.Height)

				if match, _ := regexp.MatchString("^.*\\.(MP4|mp4|mov|MOV|GIF|gif)$", fileHeader.Filename); match {
					//check for the file size
					if size := fileHeader.Size; size > 3584*MB {
						http.Error(w, "File size exceeds 3.6GB", http.StatusInternalServerError)
						return
					}

				}

				//Create a new file on the server
				//get cleaned file name
				s := regexp.MustCompile(`\s+`).ReplaceAllString(fileHeader.Filename, "")
				time := fmt.Sprintf("%v", time.Now())
				s = regexp.MustCompile(`\s+`).ReplaceAllString(time, "") + s

				dst, err := os.Create(filepath.Join("./posts", s))
				if err != nil {
					http.Error(w, "Unable to create a file", http.StatusInternalServerError)
					return
				}
				defer dst.Close()

				_, err = io.Copy(dst, file)
				if err != nil {
					http.Error(w, "Unable to write file", http.StatusInternalServerError)
					return
				}

				postPath = append(postPath, filepath.Join("./posts", s))

			}

		}
		requestBodyPostPath := fmt.Sprintf("%s", strings.Join(postPath, ","))
		insertPostPath := `UPDATE posts SET post_path=$1,complete_post=$2 WHERE post_id=$3`
		_, err = db.Query(insertPostPath, requestBodyPostPath, true, postId.PostId)
		if err != nil {
			panic(err)
			// http.Error(w, "Error inserting to DB", http.StatusInternalServerError)

		}

		json.NewEncoder(w).Encode("Media uploaded successfully")
	})

	//to get all posts of users
	http.HandleFunc("/getAllPosts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var userPosts []UsersPost

		var userId UserID
		// userId.UserId = 1
		err := json.NewDecoder(r.Body).Decode(&userId)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusMethodNotAllowed)
			return
		}
		if userId.UserId == 0 {
			fmt.Fprintln(w, "Invalid user id or missing field")
			return
		}

		getPosts := `SELECT post_id,post_path,poat_caption,location,hide_like,hide_comments,posted_on FROM posts WHERE user_id=$1 ORDER BY posted_on DESC`
		row, err := db.Query(getPosts, userId.UserId)
		if err != nil {
			panic(err)
		}
		for row.Next() {
			//to get username
			var userPost UsersPost
			getUserName := `SELECT user_name,display_pic FROM users WHERE user_id=$1`
			var url string
			err = db.QueryRow(getUserName, userId.UserId).Scan(&userPost.UserName, &url)

			userPost.UserProfilePicURL = "http://localhost:3000/getProfilePic/" + url
			if err != nil {
				http.Error(w, "Unable to get username", http.StatusInternalServerError)
				return
			}
			var postURLstr string
			err = row.Scan(&userPost.PostId, &postURLstr, &userPost.PostCaption, &userPost.AttachedLocation, &userPost.HideLikeCount, &userPost.HideLikeCount, &userPost.PostedOn)
			if err != nil {
				panic(err)
			}

			//get like status of present user
			err = db.QueryRow("SELECT EXISTS(SELECT user_name FROM likes WHERE post_id=$1 AND user_name=$2)", userPost.PostId, userPost.UserName).Scan(&userPost.LikeStatus)
			if err != nil {
				panic(err)
			}

			//get count of likes
			err = db.QueryRow("SELECT COUNT(user_name) FROM likes WHERE post_id=$1", userPost.PostId).Scan(&userPost.Likes)
			if err != nil {
				panic(err)
			}
			postURL := strings.Split(postURLstr, ",")
			for _, url := range postURL {
				url = "http://localhost:3000/download/" + url
				userPost.PostURL = append(userPost.PostURL, url)

			}

			err = db.QueryRow("SELECT user_id,post_id FROM savedposts WHERE user_id=$1.post_id=$2", userPost.UserID, userPost.PostId).Scan(&userPost.UserID, &userPost.PostId)
			if err != nil {
				userPost.SavedStatus = false
			} else {
				userPost.SavedStatus = true
			}

			userPost.UserID = userId.UserId
			userPosts = append(userPosts, userPost)
		}
		json.NewEncoder(w).Encode(userPosts)

	})

	// handle function to like 		a post
	http.HandleFunc("/likePost", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var requestBody LikePost
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusMethodNotAllowed)
			return
		}
		//validate for proper userId and PostId
		if requestBody.PostId == 0 || requestBody.UserID == 0 {
			fmt.Fprintln(w, "Invalid userId or PostId/missing fields")
			return
		}

		getUserName := `SELECT user_name FROM users WHERE user_id=$1`
		var userName string
		err = db.QueryRow(getUserName, requestBody.UserID).Scan(&userName)
		if err != nil {
			panic(err)
		}

		insertLike := `INSERT INTO likes(post_id,user_name) VALUES($1,$2)`
		_, err = db.Query(insertLike, requestBody.PostId, userName)
		if err != nil {

			_, err := db.Query("DELETE FROM likes WHERE user_name=$1", userName)
			if err != nil {
				panic(err)
			}

		}

		getTotalLikes := `SELECT COUNT(user_name) FROM likes WHERE post_id=$1`
		var likes TotalLikes
		err = db.QueryRow(getTotalLikes, requestBody.PostId).Scan(&likes.TotalLikes)
		if err != nil {
			panic(err)
		}
		json.NewEncoder(w).Encode(likes)
	})

	//handle func to comment a post based on postid
	http.HandleFunc("/commentPost", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var requestBody CommentBody
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusNotAcceptable)
			return
		}
		//validate for proper userId and PostId
		if requestBody.PostId == 0 || requestBody.UserID == 0 {
			fmt.Fprintln(w, "Invalid userId or PostId or missing fields")
			return
		}
		if requestBody.CommentBody == "" {
			http.Error(w, "CommentBody cannot be empty or missing field", http.StatusNotAcceptable)
			return
		}

		if len(requestBody.CommentBody) > 2500 {
			http.Error(w, "Comment body should not exceed 2500 characters", http.StatusNotAcceptable)
			return
		}

		insertComment := `INSERT INTO comments(commentoruser_id,post_id,comment_body) VALUES($1,$2,$3) RETURNING comment_id`
		var returnedCommentId ReturnedCommentId

		err = db.QueryRow(insertComment, requestBody.UserID, requestBody.PostId, requestBody.CommentBody).Scan(&returnedCommentId.ReturnedCommentId)
		if err != nil {
			http.Error(w, "Invalid post id", http.StatusBadRequest)
		}
		json.NewEncoder(w).Encode(returnedCommentId)

	})

	//handle func to get all comments of a post based on postId
	http.HandleFunc("/getAllComments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var postId PostId
		err := json.NewDecoder(r.Body).Decode(&postId)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusMethodNotAllowed)
			return
		}
		if postId.PostId == 0 {
			fmt.Fprintln(w, "Invalid post id")
			return
		}

		err = db.QueryRow("SELECT post_id FROM posts WHERE post_id=$1", postId.PostId).Scan(&postId.PostId)
		if err != nil {
			http.Error(w, "Invalid post id", http.StatusInternalServerError)
			return
		}

		var comments []CommentsOfPost
		var comment CommentsOfPost
		row, err := db.Query("SELECT commentoruser_id,comment_id,comment_body,commented_on FROM comments WHERE post_id=$1 ORDER BY commented_on DESC", postId.PostId)
		if err != nil {
			panic(err)
		}
		for row.Next() {
			var commentorUSerID int64
			err = row.Scan(&commentorUSerID, &comment.CommentId, &comment.CommentBody, &comment.CommentedOn)
			if err != nil {
				panic(err)
			}

			var dpURL string
			err = db.QueryRow("SELECT user_name,display_pic FROM users WHERE user_id=$1", commentorUSerID).Scan(&comment.CommentorUserName, &dpURL)
			if err != nil {
				panic(err)
			}
			comment.CommentorDisplayPic = "http://localhost:3000/getProfilePic/" + dpURL
			comment.PostId = postId.PostId
			comments = append(comments, comment)

		}
		json.NewEncoder(w).Encode(comments)

	})

	//handle function to follow(me following other)
	http.HandleFunc("/follow", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not valid", http.StatusMethodNotAllowed)
			return
		}
		var x Follow
		err := json.NewDecoder(r.Body).Decode(&x)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusMethodNotAllowed)
			return
		}

		if x.MyId == 0 || x.Following == 0 {
			fmt.Fprint(w, "Invalid IDs or missing fields")
			return
		}

		var private bool
		err = db.QueryRow("SELECT private FROM users WHERE user_id=$1", x.Following).Scan(&private)
		if err != nil {
			panic(err)
		}

		if private == true {
			_, err = db.Query("INSERT INTO follower(user_id,follower_id,accepted) VALUES($1,$2,$3)", x.MyId, x.Following, false)
			if err != nil {
				_, err = db.Query("DELETE FROM follower WHERE follower_id=$1", x.Following)
				if err != nil {
					panic(err)
				}
				fmt.Fprintln(w, "removed follow request")
				return
			}
			fmt.Fprintln(w, "Follow request pending")
			var follow FollowStatus
			follow.FollowStatus = false
			json.NewEncoder(w).Encode(follow)

		}

		if private == false {
			_, err = db.Query("INSERT INTO follower(user_id,follower_id) VALUES($1,$2)", x.MyId, x.Following)
			if err != nil {
				_, err = db.Query("DELETE FROM follower WHERE follower_id=$1", x.Following)
				if err != nil {
					panic(err)
				}
				var follow FollowStatus
				follow.FollowStatus = false
				json.NewEncoder(w).Encode(follow)
				return

			}
			var follow FollowStatus
			follow.FollowStatus = true
			json.NewEncoder(w).Encode(follow)
		}

	})

	//handle function to list followers of a user
	http.HandleFunc("/followers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method invalid", http.StatusMethodNotAllowed)
			return
		}
		var userId UserID
		err := json.NewDecoder(r.Body).Decode(&userId)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusMethodNotAllowed)
			return
		}

		var follower Follows
		var followers []Follows

		row, err := db.Query("SELECT user_id FROM follower WHERE follower_id=$1 AND accepted=$2", userId.UserId, true)
		if err != nil {
			panic(err)
		}

		for row.Next() {
			err = row.Scan(&follower.UserID)
			if err != nil {
				panic(err)
			}

			err = db.QueryRow("SELECT name,user_name,display_pic FROM users WHERE user_id=$1", follower.UserID).Scan(&follower.Name, &follower.UserName, &follower.ProfilePic)
			if err != nil {
				panic(err)
			}
			follower.ProfilePic = "http://localhost:3000/getProfilePic/" + follower.ProfilePic

			//to check following back status
			var id int64
			err = db.QueryRow("SELECT user_id FROM follower WHERE follower_id=$1 AND accepted=$2", follower.UserID, true).Scan(&id)
			if err != nil {
				follower.FollowingBackStatus = false
			}

			if id != userId.UserId {
				follower.FollowingBackStatus = false
			} else {
				follower.FollowingBackStatus = true
			}

			followers = append(followers, follower)
		}

		json.NewEncoder(w).Encode(followers)
	})

	//pending follow requests
	http.HandleFunc("/getFollowRequests", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var userId UserID
		json.NewDecoder(r.Body).Decode(&userId)

		if userId.UserId == 0 {
			http.Error(w, "Missing or invalid userId", http.StatusNotAcceptable)
			return
		}

		row, err := db.Query("SELECT user_id,created_at,accepted FROM follower WHERE follower_id=$1 AND accepted=$2", userId.UserId, false)
		if err != nil {
			panic(err)
		}
		var followRequest []FollowRequest
		for row.Next() {
			var followrequest FollowRequest
			err = row.Scan(&followrequest.UserID, &followrequest.CreatedOn, &followrequest.Accepted)
			if err != nil {
				panic(err)
			}

			err = db.QueryRow("SELECT user_name,display_pic FROM users WHERE user_id=$1", followrequest.UserID).Scan(&followrequest.UserName, &followrequest.ProfilePic)
			if err != nil {
				panic(err)
			}
			followrequest.ProfilePic = "http://localhost:3000/getProfilePhoto/" + followrequest.ProfilePic
			followRequest = append(followRequest, followrequest)

		}
		json.NewEncoder(w).Encode(followRequest)
	})

	//response to follow requests
	http.HandleFunc("/respondingRequest", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid Method", http.StatusMethodNotAllowed)
			return
		}

		var accepted FollowAcceptance
		err = json.NewDecoder(r.Body).Decode(&accepted)
		if err != nil {
			http.Error(w, "Error decoding request body", http.StatusNotAcceptable)
			return
		}

		//validate ids in db follower
		var user_id, follwer_id int64
		err = db.QueryRow("SELECT user_id,follower_id FROM follower WHERE follower_id=$1 AND accepted=$2", accepted.AcceptorUserID, false).Scan(&user_id, &follwer_id)
		if err != nil {
			http.Error(w, "Request doesn't exist", http.StatusInternalServerError)
			return
		}

		if accepted.AcceptStatus {
			_, err = db.Query("UPDATE follower SET accepted=$1 WHERE user_id=$2 AND follower_id=$3", true, accepted.RequestorId, accepted.AcceptorUserID)
			if err != nil {
				http.Error(w, "Couldn't update request", http.StatusInternalServerError)
				return
			}
			fmt.Fprintln(w, "Accepted follow request")
		} else {
			_, err = db.Query("DELETE FROM follower WHERE user_id=$1 AND follower_id=$2 AND accepted=$3", accepted.RequestorId, accepted.AcceptorUserID, false)
			if err != nil {
				http.Error(w, "Couldn't delete pending follow request", http.StatusInternalServerError)
				return
			}
			fmt.Fprintln(w, "Deleted follow request")
		}
	})

	//handleFunc to remove follower
	http.HandleFunc("/removeFollower", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var follower DeleteFollower
		err = json.NewDecoder(r.Body).Decode(&follower)
		if err != nil {
			http.Error(w, "Error decoding the request body", http.StatusBadRequest)
			return
		}

		if follower.FollowerUserId == 0 && follower.MyuserId == 0 {
			http.Error(w, "Missing or invalid Ids", http.StatusInternalServerError)
			return
		}

		_, err = db.Query("DELETE FROM follower WHERE user_id=$1 AND follower_id=$2 AND accepted=$3", follower.FollowerUserId, follower.MyuserId, true)
		if err != nil {
			http.Error(w, "Error removing the follower", http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, "Removed follower successfully")

	})

	//handle func to get list of users me following
	http.HandleFunc("/following", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method invalid", http.StatusMethodNotAllowed)
			return
		}
		var userId UserID
		err := json.NewDecoder(r.Body).Decode(&userId)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusMethodNotAllowed)
			return
		}

		row, err := db.Query("SELECT follower_id FROM follower WHERE user_id=$1 AND accepted=$2", userId.UserId, true)
		if err != nil {
			panic(err)
		}
		var following []Follows
		for row.Next() {
			var follow Follows
			err = row.Scan(&follow.UserID)
			if err != nil {
				panic(err)
			}
			err = db.QueryRow("SELECT name,user_name,display_pic FROM users WHERE user_id=$1", follow.UserID).Scan(&follow.Name, &follow.UserName, &follow.ProfilePic)
			if err != nil {
				panic(err)
			}
			follow.ProfilePic = "http://localhost:3000/getProfilePic/" + follow.ProfilePic
			follow.FollowingBackStatus = true
			following = append(following, follow)
		}
		json.NewEncoder(w).Encode(following)
	})

	//handle function to update bio in profile
	http.HandleFunc("/updateBio", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var updateProfile ProfileUpdate
		err = json.NewDecoder(r.Body).Decode(&updateProfile)
		if err != nil {
			http.Error(w, "Error decoding request", http.StatusNoContent)
			return
		}

		if updateProfile.UserID <= 0 {
			http.Error(w, "User ID not accepted or missing field", http.StatusNotAcceptable)
			return
		}

		if len(updateProfile.Bio) > 150 {
			http.Error(w, "Bio exceeds the character limit (150)", http.StatusNotAcceptable)
			return
		}

		// user name validation

		match, _ := regexp.MatchString("^[a-zA-Z0-9][a-zA-Z0-9_]*$", updateProfile.UserName)
		if !match {
			fmt.Fprintln(w, "User name should start with alphabet and can have combination minimum 8 characters of numbers and only underscore(_)")
			return
		}

		if len(updateProfile.UserName) < 7 || len(updateProfile.UserName) > 20 {
			http.Error(w, "Username should be of length(7,20)", http.StatusMethodNotAllowed)
			return
		}

		//validate name
		if len(updateProfile.Name) > 20 {
			http.Error(w, "Name should be less than 20 characters", http.StatusMethodNotAllowed)
			return
		}

		_, err = db.Query("UPDATE users SET bio =$1,name=$2,user_name=$3 WHERE user_id=$4", updateProfile.Bio, updateProfile.Name, updateProfile.UserName, updateProfile.UserID)
		if err != nil {
			panic(err)
		}

		fmt.Fprint(w, "Update successful")
	})

	//get profile
	http.HandleFunc("/userProfile", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var userId UserID
		err := json.NewDecoder(r.Body).Decode(&userId)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusMethodNotAllowed)
			return
		}

		var profile Profile
		var partialURL string

		//get info from users table
		getProfile := `SELECT user_name,display_pic,bio,private FROM users WHERE user_id=$1`
		err = db.QueryRow(getProfile, userId.UserId).Scan(&profile.UserName, &partialURL, &profile.Bio, &profile.PrivateAccount)
		if err != nil {
			panic(err)
		}

		profile.UserID = userId.UserId

		profile.ProfilePic = "http://localhost:3000/getProfilePic/" + partialURL

		//get count of total post of user
		getPostCount := `SELECT COUNT(post_id) FROM posts WHERE user_id=$1`
		err = db.QueryRow(getPostCount, userId.UserId).Scan(&profile.PostCount)
		if err != nil {
			panic(err)
		}

		//get count of followers
		getFollowerCount := `SELECT COUNT(user_id) FROM follower WHERE follower_id=$1`
		err = db.QueryRow(getFollowerCount, userId.UserId).Scan(&profile.FollowerCount)
		if err != nil {
			panic(err)
		}

		//get following count
		getFollowingCount := `SELECT COUNT(follower_id) FROM follower WHERE user_id=$1`
		err = db.QueryRow(getFollowingCount, userId.UserId).Scan(&profile.FollowingCount)
		if err != nil {
			panic(err)
		}

		json.NewEncoder(w).Encode(profile)
	})

	//to save a post
	http.HandleFunc("/savePost", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var post LikePost //reusing struct with user_id and post_id fields
		err := json.NewDecoder(r.Body).Decode(&post)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusMethodNotAllowed)
			return
		}

		if post.PostId == 0 || post.UserID == 0 {
			http.Error(w, "Invalid post or userd ID", http.StatusInternalServerError)
			return
		}

		err = db.QueryRow("SELECT post_id FROM posts WHERE post_id=$1", post.PostId).Scan(&post.PostId)
		if err != nil {
			http.Error(w, "Invalid post id/doesnt exist in db posts", http.StatusNotAcceptable)
			return
		}

		var postId, userId int64
		var savedStatus SavedStatus
		err = db.QueryRow("SELECT post_id,user_id FROM savedposts WHERE post_id=$1", post.PostId).Scan(&postId, &userId)
		if err != nil {
			//insert into savedposts  table

			_, err = db.Query("INSERT INTO savedposts(post_id,user_id) VALUES($1,$2)", post.PostId, post.UserID)
			if err != nil {
				panic(err)
			}
			fmt.Fprintln(w, "Saved successfully")
			savedStatus.SavedStatus = true
			json.NewEncoder(w).Encode(savedStatus)
			return

		}

		if postId == post.PostId && userId == post.UserID {
			_, err = db.Query("DELETE FROM savedposts WHERE post_id=$1", postId)
			if err != nil {
				panic(err)
			}
			fmt.Fprintln(w, "Removed from saved successfully")
			savedStatus.SavedStatus = false
			json.NewEncoder(w).Encode(savedStatus)
			return
		}

	})

	//handle function to get posts using post_id
	http.HandleFunc("/getpost/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		//:var id int64
		idstr := fmt.Sprint(r.URL)

		_, idstr = path.Split(idstr)

		postId, err := strconv.Atoi(idstr)
		if err != nil {
			http.Error(w, "Bad post id", http.StatusMethodNotAllowed)
			return
		}

		err = db.QueryRow("SELECT post_id FROM posts WHERE post_id=$1 AND complete_post=$2", postId, true).Scan(&postId)
		if err != nil {
			http.Error(w, "Invalid postId or does not exist", http.StatusInternalServerError)
			return
		}
		var post UsersPost

		post.PostId = int64(postId)

		var postURL string
		query := `SELECT user_id,post_path,poat_caption,location,hide_like,hide_comments,posted_on FROM posts WHERE post_id=$1 AND complete_post=$2`
		err = db.QueryRow(query, postId, true).Scan(&post.UserID, &postURL, &post.PostCaption, &post.AttachedLocation, &post.HideLikeCount, &post.TurnOffComments, &post.PostedOn)
		if err != nil {
			// http.Error(w, "Error fetching data from db posts", http.StatusInternalServerError)
			// return
			panic(err)
		}

		filetype := strings.Split(postURL, ".")
		post.FileType = GetExtension("." + filetype[len(filetype)-1])

		postURL = "http://localhost:3000/download/" + postURL
		post.PostURL = append(post.PostURL, postURL)

		var URL string
		err = db.QueryRow("SELECT user_name,display_pic FROM users WHERE user_id=$1", post.UserID).Scan(&post.UserName, &URL)
		if err != nil {
			http.Error(w, "Error retriving data from db users", http.StatusInternalServerError)
			return
		}
		post.UserProfilePicURL = "http://localhost:3000/getProfilePic/" + URL

		err = db.QueryRow("SELECT COUNT(user_name) FROM likes WHERE post_id=$1", postId).Scan(&post.Likes)
		if err != nil {
			http.Error(w, "Error retriving likes count", http.StatusInternalServerError)
			return
		}

		//pending : update like status

		err = json.NewEncoder(w).Encode(post)
		if err != nil {
			// http.Error(w,"Error encoding response",http.StatusInternalServerError)
			fmt.Fprintln(w, "Error encoding response")
			return
		}

	})

	//handle func to get all saved posts of a user
	http.HandleFunc("/savedposts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var userId UserID
		err := json.NewDecoder(r.Body).Decode(&userId)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusMethodNotAllowed)
			return
		}

		if userId.UserId == 0 {
			fmt.Fprintln(w, "Invalid User ID")
			return
		}

		row, err := db.Query("SELECT post_id FROM savedposts WHERE user_id=$1", userId.UserId)
		if err != nil {
			panic(err)
		}

		var finalpostid []SavedPosts
		for row.Next() {
			var postid SavedPosts
			var posturl string
			err = row.Scan(&postid.PostId)
			if err != nil {
				panic(err)
			}
			err = db.QueryRow("SELECT post_path FROM posts WHERE post_id=$1", postid.PostId).Scan(&posturl)
			ext := strings.ToLower(filepath.Ext(posturl))

			postid.ContentType = GetExtension(ext)
			postid.PostURL = "http://localhost:3000/download/" + posturl
			finalpostid = append(finalpostid, postid)
		}

		json.NewEncoder(w).Encode(finalpostid)

	})

	//delete post
	http.HandleFunc("/deletePost", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var delete LikePost //reusing struct
		err = json.NewDecoder(r.Body).Decode(&delete)
		if err != nil {
			http.Error(w, "Error decoding request body", http.StatusBadRequest)
			return
		}

		if delete.PostId == 0 || delete.UserID == 0 {
			fmt.Fprintln(w, "Missing or invalid ids")
			return
		}

		var postExist bool

		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM posts WHERE user_id=$1 AND post_id=$2)", delete.UserID, delete.PostId).Scan(&postExist)
		if err != nil {
			panic(err)
		}
		if !postExist {
			http.Error(w, "Invalid ids", http.StatusBadRequest)
			return
		}

		var url string
		err = db.QueryRow("SELECT post_path FROM posts WHERE post_id=$1", delete.PostId).Scan(&url)
		if err != nil {
			panic(err)
		}

		os.Remove("./" + url)

		_, err = db.Query("DELETE FROM posts WHERE post_id=$1 AND user_id=$2", delete.PostId, delete.UserID)
		if err != nil {
			panic(err)
		}

		fmt.Fprintln(w, "Post deleted successfully")

	})

	//delete user
	http.HandleFunc("/deleteAccount", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var loginCred LoginCred
		err = json.NewDecoder(r.Body).Decode(&loginCred)
		if err != nil {
			http.Error(w, "Error decoding request body", http.StatusPartialContent)
			return
		}

		if loginCred.Password == "" && loginCred.UserName == "" {
			http.Error(w, "Invalid username or Password", http.StatusPartialContent)
			return
		}

		var passwordHash string
		err = db.QueryRow("SELECT password FROM users WHERE user_name=$1", loginCred.UserName).Scan(&passwordHash)
		if err != nil {
			panic(err)
		}
		err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(loginCred.Password))
		if err == nil {
			_, err = db.Query("DELETE FROM users WHERE user_name=$1", loginCred.UserName)
			if err != nil {
				http.Error(w, "Error occured while deleting account", http.StatusInternalServerError)
				return
			}
		}
		if err != nil {
			fmt.Fprintln(w, "Invalid password")
			return
		}

	})

	//remove saved post

	http.HandleFunc("/removeSavedPost", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var remove LikePost //reusing struct
		err = json.NewDecoder(r.Body).Decode(&remove)
		if err != nil {
			http.Error(w, "Error decoding request body", http.StatusBadRequest)
			return
		}

		if remove.PostId <= 0 || remove.UserID <= 0 {
			http.Error(w, "Missing field or invalid ids", http.StatusInternalServerError)
			return
		}

		err = db.QueryRow("SELECT user_id,post_id FROM savedposts WHERE user_id=$1 AND post_id=$2", remove.UserID, remove.PostId).Scan(&remove.UserID, &remove.PostId)
		if err != nil {
			http.Error(w, "Invalid user_id", http.StatusInternalServerError)
			return
		}

		_, err = db.Query("DELETE FROM savedposts WHERE user_id=$1 AND post_id=$2", remove.UserID, remove.PostId)
		if err != nil {
			panic(err)
		}
		fmt.Fprintln(w, "Removed post from saved posts")
	})

	//turnoff commenting

	http.HandleFunc("/turnoffComments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var commentoff LikePost //reusing struct fields
		err = json.NewDecoder(r.Body).Decode(&commentoff)
		if err != nil {
			http.Error(w, "Error decoding request body", http.StatusBadRequest)
			return
		}

		if commentoff.PostId <= 0 || commentoff.UserID <= 0 {
			fmt.Fprintln(w, "Invalid ids or missing field")
			return
		}

		err = db.QueryRow("SELECT user_id,post_id FROM posts WHERE user_id=$1 AND post_id=$2", commentoff.UserID, commentoff.PostId).Scan(&commentoff.UserID, &commentoff.PostId)
		if err != nil {
			panic(err)
		}

		_, err = db.Query("UPDATE posts SET hide_comments=$1 WHERE user_id=$2", true, commentoff.UserID)
		if err != nil {
			panic(err)
		}
		fmt.Fprintln(w, "Comments turned off")

	})

	//turnon commenting

	http.HandleFunc("/turnonComments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var commentoff LikePost //reusing struct fields
		err = json.NewDecoder(r.Body).Decode(&commentoff)
		if err != nil {
			http.Error(w, "Error decoding request body", http.StatusBadRequest)
			return
		}

		if commentoff.PostId <= 0 || commentoff.UserID <= 0 {
			fmt.Fprintln(w, "Invalid ids or missing field")
			return
		}

		err = db.QueryRow("SELECT user_id,post_id FROM posts WHERE user_id=$1 AND post_id=$2", commentoff.UserID, commentoff.PostId).Scan(&commentoff.UserID, &commentoff.PostId)
		if err != nil {
			panic(err)
		}

		_, err = db.Query("UPDATE posts SET hide_comments=$1 WHERE user_id=$2", false, commentoff.UserID)
		if err != nil {
			panic(err)
		}
		fmt.Fprintln(w, "Comments turned on")

	})

	//hide like count
	http.HandleFunc("/hidelikeCount", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var commentoff LikePost //reusing struct fields
		err = json.NewDecoder(r.Body).Decode(&commentoff)
		if err != nil {
			http.Error(w, "Error decoding request body", http.StatusBadRequest)
			return
		}

		if commentoff.PostId <= 0 || commentoff.UserID <= 0 {
			fmt.Fprintln(w, "Invalid ids or missing field")
			return
		}

		err = db.QueryRow("SELECT user_id,post_id FROM posts WHERE user_id=$1 AND post_id=$2", commentoff.UserID, commentoff.PostId).Scan(&commentoff.UserID, &commentoff.PostId)
		if err != nil {
			panic(err)
		}

		_, err = db.Query("UPDATE posts SET hide_like=$1 WHERE user_id=$2", true, commentoff.UserID)
		if err != nil {
			panic(err)
		}
		fmt.Fprintln(w, "updated hide_like=true")

	})

	//show like count
	http.HandleFunc("/showlikeCount", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var commentoff LikePost //reusing struct fields
		err = json.NewDecoder(r.Body).Decode(&commentoff)
		if err != nil {
			http.Error(w, "Error decoding request body", http.StatusBadRequest)
			return
		}

		if commentoff.PostId <= 0 || commentoff.UserID <= 0 {
			fmt.Fprintln(w, "Invalid ids or missing field")
			return
		}

		err = db.QueryRow("SELECT user_id,post_id FROM posts WHERE user_id=$1 AND post_id=$2", commentoff.UserID, commentoff.PostId).Scan(&commentoff.UserID, &commentoff.PostId)
		if err != nil {
			panic(err)
		}

		_, err = db.Query("UPDATE posts SET hide_like=$1 WHERE user_id=$2", false, commentoff.UserID)
		if err != nil {
			panic(err)
		}
		fmt.Fprintln(w, "updated hide_likes=false")

	})

	//delete comment
	http.HandleFunc("/deleteComment", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var deleteComment DeleteComment
		err = json.NewDecoder(r.Body).Decode(&deleteComment)
		if err != nil {
			http.Error(w, "Error decoding request body", http.StatusBadRequest)
			return
		}

		if deleteComment.CommentId <= 0 || deleteComment.PostId <= 0 || deleteComment.UserID <= 0 {
			http.Error(w, "Missing fields or inavlid ids", http.StatusResetContent)
			return
		}

		err = db.QueryRow("SELECT user_id,post_id FROM posts WHERE user_id=$1 AND post_id=$2", deleteComment.UserID, deleteComment.PostId).Scan(&deleteComment.UserID, &deleteComment.PostId)
		if err != nil {
			http.Error(w, "Invalid Ids for operation", http.StatusInternalServerError)
			return
		}

		err = db.QueryRow("SELECT post_id,comment_id FROM comments WHERE post_id=$1 AND comment_id=$2", deleteComment.PostId, deleteComment.CommentId).Scan(&deleteComment.PostId, &deleteComment.CommentId)
		if err != nil {
			http.Error(w, "Invalid Ids for operation", http.StatusInternalServerError)
			return
		}

		_, err = db.Query("DELETE FROM comments WHERE post_id=$1 AND comment_id=$2", deleteComment.PostId, deleteComment.CommentId)
		if err != nil {
			http.Error(w, "error deleting comment", http.StatusInternalServerError)
			return
		}

		fmt.Fprintln(w, "Comment deleted succefully")

	})

	//api for searching users

	http.HandleFunc("/searchAccounts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var username UserName
		err = json.NewDecoder(r.Body).Decode(&username)
		if err != nil {
			http.Error(w, "error decoding request body", http.StatusBadRequest)
			return
		}

		if username.UserName == "" {
			http.Error(w, "Invalid user name or missing field", http.StatusPartialContent)
			return
		}

		str := regexp.MustCompile(`[a-zA-Z]*`)
		name := str.FindAllString(username.UserName, 1)

		num := regexp.MustCompile(`\d+`)
		number := num.FindAllString(username.UserName, 1)

		var like string
		if len(number) == 0 {
			like = "%" + name[0] + "%"
		}
		if len(name) == 0 {
			like = "%" + number[0] + "%"
		}
		if len(name) != 0 && len(number) != 0 {
			like = "%" + name[0] + "%" + number[0] + "%"
		}
		row, err := db.Query("SELECT user_id,user_name,name,display_pic FROM users WHERE user_name ILIKE $1", like)
		if err != nil {
			panic(err)
		}

		var accounts []Accounts
		for row.Next() {
			var acc Accounts
			err = row.Scan(&acc.UserID, &acc.UserName, &acc.Name, &acc.ProfilePic)
			if err != nil {
				panic(err)
			}
			acc.ProfilePic = "http://localhost:3000/getProfilePic/" + acc.ProfilePic
			accounts = append(accounts, acc)

		}

		json.NewEncoder(w).Encode(accounts)
	})

	//search for hashtags
	http.HandleFunc("/searchHashtag", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var hashtag HashtagSearch
		err = json.NewDecoder(r.Body).Decode(&hashtag)
		if err != nil {
			http.Error(w, "Error decoding request body", http.StatusPartialContent)
			return
		}
		str := regexp.MustCompile(`[a-zA-Z_]*`)
		name := str.FindAllString(hashtag.Hashtag, 1)

		num := regexp.MustCompile(`\d+`)
		number := num.FindAllString(hashtag.Hashtag, 1)

		var like string
		if len(number) == 0 {
			like = name[0] + "%"
		}
		if len(name) == 0 {
			like = number[0] + "%"
		}
		if len(name) != 0 && len(number) != 0 {
			like = name[0] + "%" + number[0] + "%"
		}

		row, err := db.Query("SELECT hash_id,hash_name FROM hashtags WHERE hash_name ILIKE $1", like)
		if err != nil {
			panic(err)
		}

		var results []HashtagSearchResult
		for row.Next() {
			var result HashtagSearchResult
			err = row.Scan(&result.HashId, &result.HashName)
			if err != nil {
				http.Error(w, "error reading hashtable", http.StatusInternalServerError)
				return
			}

			if result.HashId == 0 {
				fmt.Println("its empty")

			}

			err = db.QueryRow("SELECT COUNT(post_id) FROM mentions WHERE hash_id=$1", result.HashId).Scan(&result.PostCount)
			if err != nil {
				http.Error(w, "Error getting count of posts of hash", http.StatusInternalServerError)
				return
			}

			results = append(results, result)

		}
		var newhashtag Newhashtag
		if len(results) == 0 {

			err = db.QueryRow("INSERT INTO hashtags(hash_name) VALUES($1) RETURNING hash_id", hashtag.Hashtag).Scan(&newhashtag.NewHashId)
			if err != nil {
				panic(err)
				// http.Error(w, "Error creating new hash", http.StatusInternalServerError)
				// return
			}

			json.NewEncoder(w).Encode(newhashtag)

		}

		if len(results) != 0 {
			json.NewEncoder(w).Encode(results)
		}

	})

	//handle func to upload story info
	http.HandleFunc("/uploadStoryInfo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var storyinfo StoryInfo
		err = json.NewDecoder(r.Body).Decode(&storyinfo)
		if err != nil {
			http.Error(w, "Error decoding request body", http.StatusBadRequest)
			return
		}

		if len(storyinfo.TaggedIds) > 20 {
			http.Error(w, "Maximum 20 ids allowed", http.StatusBadRequest)
			return
		}
		var idexists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE user_id=$1)", storyinfo.UserID).Scan(&idexists)
		if err != nil {
			http.Error(w, "Invalid user-id", http.StatusBadRequest)
			return
		}
		if !idexists {
			http.Error(w, "Invalid user-id", http.StatusBadRequest)
			return
		}

		var storyIds []ReturnedStoryId

		//check for existance of all tagged ids
		for _, ids := range storyinfo.TaggedIds {

			var returnedStoryId ReturnedStoryId
			err = db.QueryRow("INSERT INTO stories(user_id,story_path) VALUES($1,$2) RETURNING story_id", storyinfo.UserID, "").Scan(&returnedStoryId.ReturnedStoryId)
			if err != nil {
				panic(err)
			}

			returnedStoryId.PostAsStory = false
			storyIds = append(storyIds, returnedStoryId)

			for _, id := range ids {
				var count int64
				err = db.QueryRow("SELECT COUNT(story_id) FROM stories WHERE user_id=$1", storyinfo.UserID).Scan(&count)
				if err != nil {
					http.Error(w, "Error retrieving count of stories of a user", http.StatusInternalServerError)
					return
				}

				if count < 100 {

					err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE user_id=$1)", id).Scan(&idexists)
					if err != nil {
						http.Error(w, "Invalid user-id", http.StatusInternalServerError)
						return
					}
					if !idexists {
						http.Error(w, "No user exists with this id:", http.StatusInternalServerError)
						fmt.Fprint(w, id)
						return
					}
					_, err = db.Query("INSERT INTO story_tags(story_id,tagged_id) VALUES($1,$2)", returnedStoryId.ReturnedStoryId, id)
					if err != nil {
						log.Panic(err)
						// http.Error(w, "Error inserting to story_tags table", http.StatusInternalServerError)
						// return
					}
				} else {
					_, err = db.Query("DELETE FROM stories WHERE story_id=(SELECT MIN(story_id) WHERE user_id=$1)", storyinfo.UserID)
					if err != nil {
						http.Error(w, "Coudn't delete initial post ", http.StatusInternalServerError)
						return
					}
					err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE user_id=$1)", id).Scan(&idexists)
					if err != nil {
						http.Error(w, "Invalid user-id", http.StatusInternalServerError)
						return
					}
					if !idexists {
						http.Error(w, "No user exists with this id:", http.StatusInternalServerError)
						fmt.Fprint(w, id)
						return
					}
					_, err = db.Query("INSERT INTO story_tags(story_id,tagged_id) VALUES($1,$2)", returnedStoryId, id)
					if err != nil {
						http.Error(w, "Error inserting to story_tags table", http.StatusInternalServerError)
						return
					}
				}

			}
		}
		json.NewEncoder(w).Encode(storyIds)

	})

	//handle func to upload story media
	http.HandleFunc("/uploadStoryPath", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1024*MB)
		err := r.ParseMultipartForm(1024 * MB)
		if err != nil {
			http.Error(w, "Error parsing multipart form data or file size may be out of bound", http.StatusInternalServerError)
			return
		}

		jsonData := r.FormValue("storyId")

		var storyinfo StoryMedia

		err = json.Unmarshal([]byte(jsonData), &storyinfo)
		if err != nil {
			http.Error(w, "Error unmarshalling JSON data", http.StatusBadRequest)
			return
		}

		//validate storyId
		var storyExists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM stories WHERE story_id=$1)", storyinfo.StoryId).Scan(&storyExists)
		if err != nil {
			http.Error(w, "Invalid storyId", http.StatusBadRequest)
			return
		}

		file, fileHeader, err := r.FormFile("media")
		if err != nil {
			http.Error(w, "Missing formfile", http.StatusNoContent)
			return
		}

		//get cleaned file name
		s := regexp.MustCompile(`\s+`).ReplaceAllString(fileHeader.Filename, "")
		time := fmt.Sprintf("%v", time.Now())
		s = regexp.MustCompile(`\s+`).ReplaceAllString(time, "") + s

		file, err = fileHeader.Open()
		if err != nil {
			http.Error(w, "Unable to open the file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		//check for file allowed file format
		match, _ := regexp.MatchString("^.*\\.(jpg|JPG|png|PNG|JPEG|jpeg|bmp|BMP|MP4|mp4|mov|MOV|GIF|gif)$", fileHeader.Filename)
		if !match {
			fmt.Fprintln(w, "Only JPG,JPEG,PNG,BMP formats are allowed for upload")
			return
		} else {
			//check for the file size
			if size := fileHeader.Size; size > 30*MB {
				http.Error(w, "File size exceeds 30MB", http.StatusInternalServerError)
				return
			}
		}

		if match, _ := regexp.MatchString("^.*\\.(MP4|mp4|mov|MOV|GIF|gif)$", fileHeader.Filename); match {
			//check for the file size
			if size := fileHeader.Size; size > 1024*MB {
				http.Error(w, "File size exceeds 1GB", http.StatusInternalServerError)
				return
			}

		}

		dst, err := os.Create(filepath.Join("./stories", s))
		if err != nil {
			http.Error(w, "Unable to create a file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			http.Error(w, "Unable to write file", http.StatusInternalServerError)
			return
		}
		storyPath := "stories/" + s

		var upload UploadStory
		err = db.QueryRow("UPDATE stories SET story_path=$1,success=$2 WHERE story_id=$3 RETURNING story_id,success", storyPath, true, storyinfo.StoryId).Scan(&upload.StoryId, &upload.Uploaded)
		if err != nil {
			http.Error(w, "Invalid id", http.StatusBadRequest)
			delete := "./" + storyPath
			os.Remove(delete)

			// http.Error(w, "Error inserting story media", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(upload)

	})

	//download story

	http.HandleFunc("/getStory", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var storyid StoryMedia
		err = json.NewDecoder(r.Body).Decode(&storyid)
		if err != nil {
			http.Error(w, "Error decoding request body", http.StatusBadRequest)
			return
		}

		if storyid.StoryId <= 0 {
			http.Error(w, "Invalid id or missing field", http.StatusBadRequest)
			return
		}

		//validate storyId
		var storyExists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM stories WHERE story_id=$1)", storyid.StoryId).Scan(&storyExists)
		if err != nil {
			http.Error(w, "Invalid storyId", http.StatusBadRequest)
			return
		}

		if !storyExists {
			http.Error(w, "Invalid storyId", http.StatusBadRequest)
			return
		}

		var getstory GetStory
		err = db.QueryRow("SELECT story_id,story_path,posted_on,success FROM stories WHERE story_id=$1", storyid.StoryId).Scan(&getstory.StoryId, &getstory.StoryURL, &getstory.PostedOn, &getstory.Success)
		if err != nil {
			panic(err)
		}

		row, err := db.Query("SELECT tagged_id FROM story_tags WHERE story_id=$1", storyid.StoryId)
		if err != nil {
			http.Error(w, "Error getting tagged ids", http.StatusInternalServerError)
			return
		}

		for row.Next() {
			var tagged_ids int64
			err = row.Scan(&tagged_ids)
			if err != nil {
				http.Error(w, "Scan error on tagged_id", http.StatusInternalServerError)
				return
			}

			getstory.TaggedIds = append(getstory.TaggedIds, tagged_ids)

		}
		filetype := strings.Split(getstory.StoryURL, ".")
		getstory.FileType = GetExtension("." + filetype[len(filetype)-1])
		getstory.StoryURL = "http://localhost:3000/download/" + getstory.StoryURL

		json.NewEncoder(w).Encode(getstory)

	})

	//download story api
	//handle func to serve posts
	http.HandleFunc("/download/stories/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		url := fmt.Sprint(r.URL)

		_, file := path.Split(url)

		imagePath := "./stories/" + file

		imagedata, err := ioutil.ReadFile(imagePath)
		if err != nil {
			http.Error(w, "Couldn't read the file", http.StatusInternalServerError)
			return
		}

		ext := strings.ToLower(filepath.Ext(file))

		contentType := GetExtension(ext)

		if contentType == "" {
			http.Error(w, "Unsupported file format", http.StatusUnsupportedMediaType)
			return
		}

		w.Header().Set("Content-Type", contentType)

		_, err = w.Write(imagedata)
		if err != nil {
			http.Error(w, "failed to write image data to response", http.StatusInternalServerError)
			return
		}
	})

	//delete story
	http.HandleFunc("/deleteStory", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var storyid StoryMedia
		err = json.NewDecoder(r.Body).Decode(&storyid)
		if err != nil {
			http.Error(w, "Error decoding request body", http.StatusBadRequest)
			return
		}

		if storyid.StoryId <= 0 {
			http.Error(w, "Invalid id or missing field", http.StatusBadRequest)
			return
		}

		var storypath string
		err = db.QueryRow("SELECT story_path FROM stories WHERE story_id=$1", storyid.StoryId).Scan(&storypath)
		if err != nil {
			http.Error(w, "Invalid id", http.StatusBadRequest)
			return
		}

		filelocation := "./" + storypath

		os.Remove(filelocation)

		_, err = db.Query("DELETE FROM stories WHERE story_id=$1", storyid.StoryId)
		if err != nil {
			http.Error(w, "Error deleting story", http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, "Deleted successfully")

	})

	//check post upload status
	http.HandleFunc("/postUploadStatus", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not allowed", http.StatusMethodNotAllowed)
			return
		}
		var post_id PostId
		err = json.NewDecoder(r.Body).Decode(&post_id)
		if err != nil {
			panic(err)
		}
		var postUploadStatus SavedStatus
		err = db.QueryRow("SELECT complete_post FROM posts WHERE post_id=$1", post_id.PostId).Scan(&postUploadStatus.SavedStatus)
		if err != nil {
			panic(err)
		}

		json.NewEncoder(w).Encode(postUploadStatus)
	})

	//check story upload status
	http.HandleFunc("/storyUploadStatus", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not allowed", http.StatusMethodNotAllowed)
			return
		}
		var story_id PostId
		err = json.NewDecoder(r.Body).Decode(&story_id)
		if err != nil {
			panic(err)
		}
		var postUploadStatus SavedStatus
		err = db.QueryRow("SELECT success FROM stories WHERE story_id=$1", story_id.PostId).Scan(&postUploadStatus.SavedStatus)
		if err != nil {
			http.Error(w, "Invalid story id", http.StatusBadRequest)
			return
		}

		json.NewEncoder(w).Encode(postUploadStatus)
	})

	//get active stories for a user
	http.HandleFunc("/getActiveStories", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not allowed", http.StatusMethodNotAllowed)
			return
		}
		var userId UserID
		err := json.NewDecoder(r.Body).Decode(&userId)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusMethodNotAllowed)
			return
		}

		var following []int64
		row, err := db.Query("SELECT follower_id FROM follower WHERE user_id=$1 AND accepted=$2", userId.UserId, true)
		if err != nil {
			log.Panicln("no follwings", err)
		}
		for row.Next() {
			var id int64
			err = row.Scan(&id)
			if err == sql.ErrNoRows {
				fmt.Fprintln(w, "null")
				return
			}
			following = append(following, id)
		}

		var activeStory []ActiveStories
		for _, id := range following {
			var story ActiveStories
			row, err := db.Query("SELECT story_id FROM stories WHERE user_id=$1 AND success =$2", id, true)
			if err != nil {
				panic(err)
			}
			err = db.QueryRow("SELECT user_name,display_pic FROM users WHERE user_id=$1", id).Scan(&story.User_name, &story.Profile_picURL)
			if err != nil {
				log.Panicln(err)
			}
			story.Profile_picURL = "http://localhost:3000/getProfilePic/profilePhoto/" + story.Profile_picURL

			story.User_id = id
			for row.Next() {
				var story_id int64
				err = row.Scan(&story_id)
				if err != nil {
					log.Panicln("no story id")
				}
				story.Story_id = append(story.Story_id, story_id)
				err = db.QueryRow("SELECT seen_status FROM story_seen_status WHERE user_id=$1 AND story_id=$2", userId.UserId, story_id).Scan(&story.Seen_status)
				if err != nil {
					//log.Panicln("no seen status", err)
					continue
				}

			}
			activeStory = append(activeStory, story)
		}
		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(activeStory)
	})

	//updates story seen status
	http.HandleFunc("/updateStorySeenStatus", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Method Not allowed", http.StatusMethodNotAllowed)
			return
		}
		var updateStory UpdateStorySeenStatus
		err = json.NewDecoder(r.Body).Decode(&updateStory)
		if err != nil {
			log.Panicln(err)
		}
		var exist bool
		err = db.QueryRow("SELECT EXISTS(SELECT user_id,story_id FROM story_seen_status WHERE user_id=$1 AND story_id=$2)", updateStory.UserID, updateStory.StoryId).Scan(&exist)
		if err != nil {
			log.Panicln(err)
		}

		if exist {
			return
		}
		_, err = db.Query("INSERT INTO story_seen_status(user_id,story_id,seen_status) VALUES($1,$2,$3)", updateStory.UserID, updateStory.StoryId, true)
		if err != nil {
			log.Panicln(err)
		}
	})

	http.ListenAndServe(":3000", nil)

}