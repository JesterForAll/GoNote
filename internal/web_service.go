package internal

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
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
	Correct     bool   `json:"correct"`
	CorrectNote string `json:"correctNote"`
}

type confirmRequest struct {
	Note          string `json:"note"`
	Octave        string `json:"octave"`
	CurrentNote   string `json:"currentNote"`
	CurrentOctave string `json:"currentOctave"`
}

func getMainHandle() http.Handler {
	data, err := os.ReadFile("../static/index.html")

	if err != nil {
		panic(err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(data)
	})

	return handler
}

var ListNotes = availibleNotes{
	Octaves: []string{"-4", "-3", "-2", "-1", "1", "2", "3", "4", "5"},
	Notes:   []string{"до, C", "до#, C#", "ре, D", "ре#, D#", "ми, E", "фа, F", "фа#, F#", "соль, G", "соль#, G#", "ля, A", "ля#, A#", "си, B"},
}

func getAvailibleNotes() http.Handler {

	data, err := json.Marshal(ListNotes)

	if err != nil {
		panic(err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(data)
	})

	return handler
}

func getNewNote() http.Handler {

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fileList, err := os.ReadDir("../assets")

		if err != nil {
			panic(err)
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
			panic(err)
		}

		log.Printf("Отправлен ответ: %s", string(data))

		w.Write(data)
	})

	return handler
}

func getConfirm() http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var confirmRequest confirmRequest

		err := json.NewDecoder(r.Body).Decode(&confirmRequest)

		if err != nil {
			panic(err)
		}

		log.Println("got input", confirmRequest)

		var confRes confirmResponse

		if confirmRequest.CurrentNote == confirmRequest.Note && confirmRequest.CurrentOctave == confirmRequest.Octave {
			confRes.Correct = true
		} else {
			confRes.Correct = false
			confRes.CorrectNote = confirmRequest.CurrentOctave + ", " + confirmRequest.CurrentNote
		}

		data, err := json.Marshal(confRes)

		if err != nil {
			panic(err)
		}

		log.Printf("response %s\n", string(data))

		w.Write(data)
	})

	return handler
}

func CreateServer() *http.ServeMux {
	serv := http.NewServeMux()

	serv.Handle("/", getMainHandle())
	serv.Handle("/api/availibleNotes", getAvailibleNotes())
	serv.Handle("/api/confirm", getConfirm())
	serv.Handle("/api/new-note", getNewNote())
	// serv.Handle("/api/prev-note", nil)
	serv.Handle("/notes/", http.StripPrefix("/notes/", http.FileServer(http.Dir("../assets"))))

	return serv
}
