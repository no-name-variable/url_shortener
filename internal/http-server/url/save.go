package save

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	"url_shortner/internal/lib/api/response"
	"url_shortner/internal/lib/random"
	"url_shortner/internal/storage"
	"url_shortner/internal/storage/lib/logger/sl"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	response.Response
	Alias string `json:"alias,omitempty"`
}

const aliasLength = 6

type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

func New(log *slog.Logger, urlSave URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.http.save.New"
		log = log.With(
			slog.String("op", op))
		slog.String("request_id", middleware.GetReqID(r.Context()))

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, response.Error("failed to decode request"))
			return
		}
		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)

			log.Error("invalid request", sl.Err(err))
			render.JSON(w, r, response.ValidationError(validateErr))
			return
		}

		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLength)
		}

		_, err = urlSave.SaveURL(req.URL, alias)
		if errors.Is(err, storage.ErrURLExists) {
			var errMsg = "url already exists"
			log.Info(errMsg, slog.String("url", req.URL))
			render.JSON(w, r, response.Error(errMsg))
			return
		}
		if err != nil {
			if errors.Is(err, storage.ErrURLExists) {
				var errMsg = "url already exists"
				log.Info(errMsg, slog.String("url", req.URL))
				render.JSON(w, r, response.Error(errMsg))
				return
			}
			log.Error("failed to add url", sl.Err(err))
			render.JSON(w, r, response.Error("failed to add url"))
			return
		}
		render.JSON(w, r, Response{
			response.OK(),
			alias,
		})

	}
}
