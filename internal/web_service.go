package internal

import (
	"encoding/json"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type availibleNotes struct {
	Octaves []string `json:"octaves"`
	Notes   []string `json:"notes"`
}

type note struct {
	Note     string `json:"note"`
	Octave   string `json:"octave"`
	Index    int    `json:"index"`
	AudioUrl string `json:"audioUrl"`
}

type confirmResponse struct {
	Correct     bool    `json:"correct"`
	CorrectNote string  `json:"correctNote"`
	Accuracy    float32 `json:"accuracy"`
}

type confirmRequest struct {
	Note          string `json:"note"`
	Octave        string `json:"octave"`
	CurrentNote   string `json:"currentNote"`
	CurrentOctave string `json:"currentOctave"`
}

type MainServ struct {
	Serv   *http.ServeMux
	Logger *slog.Logger
	Db     *gorm.DB
}

type dbStruct struct {
	gorm.Model
	Note         string
	Octave       string
	NumTries     int
	CorrectCount int
	Accuracy     float32
}

var ListNotes = availibleNotes{
	Octaves: []string{"-4", "-3", "-2", "-1", "1", "2", "3", "4", "5"},
	Notes:   []string{"до, C", "до#, C#", "ре, D", "ре#, D#", "ми, E", "фа, F", "фа#, F#", "соль, G", "соль#, G#", "ля, A", "ля#, A#", "си, B"},
}

func NewServer(logger *slog.Logger) *MainServ {
	serv := http.NewServeMux()

	db, err := gorm.Open(sqlite.Open("../../static/accuracy.db"), &gorm.Config{})
	if err != nil {
		logger.Error("failed to connect database", slog.Any("err", err))
	}

	db.AutoMigrate(&dbStruct{})

	mServ := &MainServ{
		Serv:   serv,
		Logger: logger,
		Db:     db,
	}

	serv.HandleFunc("GET /", mServ.handleGetMain)
	serv.HandleFunc("GET /api/availibleNotes", mServ.handleGetAvailibleNotes)
	serv.HandleFunc("POST /api/confirm", mServ.handlePostConfirm)
	serv.HandleFunc("GET /api/new-note", mServ.handleGetNextNote)
	// serv.Handle("/api/prev-note", nil)
	serv.Handle("GET /notes/", http.StripPrefix("/notes/", http.FileServer(http.Dir("../../assets"))))

	return mServ
}

func (mServ *MainServ) handleGetMain(w http.ResponseWriter, _ *http.Request) {

	data, err := os.ReadFile("../../static/index.html")

	if err != nil {
		mServ.Logger.Error("error reading main page", slog.Any("err", err))
		http.Error(w, "Internal server error while reading main page", http.StatusInternalServerError)

		return
	}

	w.Write(data)

}

func (mServ *MainServ) handleGetAvailibleNotes(w http.ResponseWriter, _ *http.Request) {

	data, err := json.Marshal(ListNotes)

	if err != nil {
		mServ.Logger.Error("error encoding data", slog.Any("err", err))
		http.Error(w, "Internal server error while encoding data", http.StatusInternalServerError)

		return
	}

	w.Write(data)

}

func (mServ *MainServ) handleGetNextNote(w http.ResponseWriter, _ *http.Request) {

	fileList, err := os.ReadDir("../../assets")

	if err != nil {
		mServ.Logger.Error("error getting data from disk", slog.Any("err", err))
		http.Error(w, "Internal server error while getting data from disk", http.StatusInternalServerError)

		return
	}

	newR := rand.New(rand.NewSource(time.Now().UnixNano()))
	randNote := fileList[newR.Intn(len(fileList))]

	// note name is in format "1-я октава (1), ре, D.wav"
	var NoteForSrv note

	indxLead := strings.Index(randNote.Name(), "(")
	indxEnd := strings.Index(randNote.Name(), ")")

	if indxLead != -1 && indxEnd != -1 {
		NoteForSrv.Octave = randNote.Name()[indxLead+1 : indxEnd]
	}

	indxDot := strings.Index(randNote.Name(), ".")
	if indxDot != -1 {
		NoteForSrv.Note = strings.TrimSpace(randNote.Name()[indxEnd+2 : indxDot])
	}

	NoteForSrv.AudioUrl = randNote.Name() //"http://localhost:9090/notes/" +

	data, err := json.Marshal(NoteForSrv)

	if err != nil {
		mServ.Logger.Error("error encoding response", slog.Any("err", err))
		http.Error(w, "Internal server error while encoding response", http.StatusInternalServerError)

		return
	}

	mServ.Logger.Info("Отправлен ответ: \n", "data", data)

	w.Write(data)

}

func (mServ *MainServ) handlePostConfirm(w http.ResponseWriter, r *http.Request) {

	var confirmRequest confirmRequest

	err := json.NewDecoder(r.Body).Decode(&confirmRequest)

	if err != nil {
		mServ.Logger.Error("error decoding request", slog.Any("err", err))
		http.Error(w, "Internal server error while decoding body", http.StatusInternalServerError)

		return
	}

	mServ.Logger.Info("got input\n", "confirmRequest", confirmRequest)

	var confRes confirmResponse
	var noteData *dbStruct

	res := mServ.Db.Where("Note = ? AND Octave = ?", confirmRequest.CurrentNote, confirmRequest.CurrentOctave).First(&noteData)

	if res.RowsAffected == 0 {
		noteData.Note = confirmRequest.CurrentNote
		noteData.Octave = confirmRequest.CurrentOctave
	}

	if confirmRequest.CurrentNote == confirmRequest.Note && confirmRequest.CurrentOctave == confirmRequest.Octave {
		confRes.Correct = true
		noteData.CorrectCount++
	} else {
		confRes.Correct = false
		confRes.CorrectNote = confirmRequest.CurrentOctave + ", " + confirmRequest.CurrentNote
	}

	noteData.NumTries++
	noteData.Accuracy = float32(noteData.CorrectCount) / float32(noteData.NumTries) * 100.00

	//update data in db
	mServ.Db.Save(noteData)

	confRes.Accuracy = noteData.Accuracy

	data, err := json.Marshal(confRes)

	if err != nil {
		mServ.Logger.Error("error encoding response", slog.Any("err", err))
		http.Error(w, "Internal server error while encoding response", http.StatusInternalServerError)

		return
	}

	mServ.Logger.Info("response\n", "data", data)

	w.Write(data)

}
