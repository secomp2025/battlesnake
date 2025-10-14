package handlers

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/secomp2025/localsnake/controllers"
	"github.com/secomp2025/localsnake/database"
)

const uploadPath = "uploads"

// UploadSnake handles the (mocked) snake file upload via HTMX multipart form.
// It returns a small HTML fragment indicating the result.
func UploadSnake(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("upload: method not allowed: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, "Method not allowed")
		return
	}

	// Ensure UTF-8 is used for any response body/fragments
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Validate session
	teamCode := GetCookieValue(r, "team_code")
	if teamCode == "" {
		log.Printf("upload: missing team_code cookie from %s", r.RemoteAddr)
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Sessão inválida. Faça login novamente.")
		return
	}

	// Validate code/team exists
	codes := controllers.NewCodeController(database.DB)
	code, err := codes.FindCode(r.Context(), teamCode)
	if err != nil || code == nil {
		log.Printf("upload: invalid code: team_code=%q err=%v", teamCode, err)
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Código inválido.")
		return
	}
	teams := controllers.NewTeamController(database.DB)
	team, err := teams.GetTeamByCode(r.Context(), code.ID)
	if err != nil || team == nil {
		log.Printf("upload: team not found for code_id=%d team_code=%q err=%v", code.ID, teamCode, err)
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Time inválido.")
		return
	}

	// Limit request body to 2MB to avoid large uploads
	const maxUpload = 2 << 20 // 2 MiB
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)
	if err := r.ParseMultipartForm(maxUpload); err != nil {
		log.Printf("upload: ParseMultipartForm failed: team_code=%q err=%v", teamCode, err)
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		fmt.Fprint(w, "Arquivo excede o limite de 2MB")
		return
	}

	file, header, err := r.FormFile("snake")
	if err != nil {
		log.Printf("upload: missing form file 'snake': team_code=%q err=%v", teamCode, err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Arquivo não recebido")
		return
	}
	defer file.Close()

	// Validate extension
	validExt := map[string]bool{".py": true, ".js": true, ".c": true}
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !validExt[ext] {
		log.Printf("upload: invalid extension: filename=%q ext=%q team_code=%q", header.Filename, ext, teamCode)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Formato inválido. Envie um arquivo .py, .js ou .c")
		return
	}

	dstPath, shouldReturn := saveSnakeFile(teamCode, teamCode, w, ext, file, header)
	if shouldReturn {
		return
	}

	// Update snake in database
	log.Println("upload: creating controller")
	snakes := controllers.NewSnakeController(database.DB)
	log.Println("upload: listing team snakes")
	team_snakes, err := snakes.ListTeamSnakes(r.Context(), team.ID)
	if err != nil {
		log.Printf("upload: failed to list snakes: team_code=%q err=%v", teamCode, err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Falha ao listar snakes")
		return
	}

	snake_server_manager := controllers.GetServerManager()

	var snake_model *database.Snake
	if len(team_snakes) > 1 {
		snake_server_manager.StopAndRemoveSnake(team_snakes[0].ID)

		original := team_snakes[0]
		original.Path = dstPath
		original.Lang = ext
		if snake_model, err = snakes.UpdateSnake(r.Context(), &original); err != nil {
			log.Printf("upload: failed to update snake: team_code=%q err=%v", teamCode, err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Falha ao atualizar snake")
			return
		}
	} else {
		log.Println("upload: creating snake")
		snake_model, err = snakes.CreateSnake(r.Context(), &database.Snake{
			TeamID: team.ID,
			Path:   dstPath,
			Lang:   ext,
		})
		if err != nil {
			log.Printf("upload: failed to create snake: team_code=%q err=%v", teamCode, err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Falha ao criar snake")
			return
		}
	}
	log.Println("upload: managing snake")
	snake_server_manager.ManageSnake(r.Context(), snake_model)

	// Success: send ASCII-only HX-Trigger header (Unicode escaped) to avoid mojibake in headers.
	SetHeader(w, "HX-Trigger", `{"show-toast": {"message": "Upload conclu\u00EDdo com sucesso.", "type": "success"}}`)
	fmt.Fprintf(w, "<div class=\"text-emerald-700\">Arquivo salvo</div>")
}

func saveSnakeFile(teamCod string, teamCode string, w http.ResponseWriter, ext string, file multipart.File, header *multipart.FileHeader) (string, bool) {
	dir := filepath.Join(uploadPath, teamCod)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Printf("upload: MkdirAll failed: dir=%q team_code=%q err=%v", dir, teamCode, err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Não foi possível criar diretório de uploads")
		return "", true
	}

	// Destination path is snake.<ext>
	dstPath := filepath.Join(dir, "snake"+ext)

	// If a current snake already exists (any allowed ext), move it to snake_prev.<ext>
	// Keep only one previous version.
	for _, e := range []string{".py", ".js", ".c"} {
		old := filepath.Join(dir, "snake"+e)
		if _, statErr := os.Stat(old); statErr == nil {
			prev := filepath.Join(dir, "snake_prev"+e)
			// Remove existing previous if present, then move current to previous
			if remErr := os.Remove(prev); remErr == nil {
				log.Printf("upload: removed existing previous: %s", prev)
			}
			if renErr := os.Rename(old, prev); renErr != nil {
				log.Printf("upload: failed to move current to previous: old=%s prev=%s err=%v", old, prev, renErr)
			} else {
				log.Printf("upload: moved current to previous: %s -> %s", old, prev)
			}
		}
	}

	// Save file atomically: write to temp, then rename
	tmpFile, err := os.CreateTemp(dir, "snake-*.tmp")
	if err != nil {
		log.Printf("upload: CreateTemp failed: dir=%q err=%v", dir, err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Falha ao preparar arquivo temporário")
		return "", true
	}
	tmpName := tmpFile.Name()
	defer func() {
		tmpFile.Close()
		// Best-effort cleanup on error
		if _, statErr := os.Stat(dstPath); errors.Is(statErr, os.ErrNotExist) {
			if rmErr := os.Remove(tmpName); rmErr == nil {
				log.Printf("upload: cleaned temp file: %s", tmpName)
			}
		}
	}()

	if _, err := io.Copy(tmpFile, file); err != nil {
		log.Printf("upload: io.Copy failed: tmp=%q filename=%q team_code=%q err=%v", tmpName, header.Filename, teamCode, err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Falha ao salvar arquivo")
		return "", true
	}
	if err := tmpFile.Sync(); err != nil {
		log.Printf("upload: tmpFile.Sync failed: tmp=%q err=%v", tmpName, err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Falha ao sincronizar arquivo")
		return "", true
	}
	if err := tmpFile.Close(); err != nil {
		log.Printf("upload: tmpFile.Close failed: tmp=%q err=%v", tmpName, err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Falha ao finalizar arquivo")
		return "", true
	}
	if err := os.Rename(tmpName, dstPath); err != nil {
		log.Printf("upload: rename failed: tmp=%q dst=%q err=%v", tmpName, dstPath, err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Falha ao mover arquivo para destino")
		return "", true
	}
	return dstPath, false
}
