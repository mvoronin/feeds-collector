package server

import (
	"encoding/json"
	"feedscollector/internal/models"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func (api *API) ListChannels(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	channels, err := api.Queries.ListAllFeedChannel(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(channels)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// AddChannel handles POST requests to add a new channel
func (api *API) AddChannel(w http.ResponseWriter, r *http.Request) {
	var params models.CreateFeedChannelParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := ValidateStruct(w, &params); err != nil {
		return
	}
	ctx := r.Context()
	if _, err := api.Queries.CreateFeedChannel(ctx, params); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (api *API) UpdateChannel(w http.ResponseWriter, r *http.Request) {
	var params models.UpdateFeedChannelParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := ValidateStruct(w, &params); err != nil {
		return
	}
	ctx := r.Context()
	if err := api.Queries.UpdateFeedChannel(ctx, params); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DeleteChannel handles DELETE requests to delete a channel
func (api *API) DeleteChannel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	if err := api.Queries.DeleteFeedChannel(ctx, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListItems handles GET requests to list all items of a channel
func (api *API) ListItems(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	items, err := api.Queries.ListFeedItem(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(items)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (api *API) RemoveItemFromChannel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channelId, err := strconv.ParseInt(vars["channel_id"], 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	itemId, err := strconv.ParseInt(vars["item_id"], 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	params := models.RemoveFeedItemFromChannelParams{
		ChannelID: channelId,
		ItemID:    itemId,
	}
	if err := api.Queries.RemoveFeedItemFromChannel(ctx, params); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DeleteItem handles DELETE requests to delete an item
func (api *API) DeleteItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	if err := api.Queries.DeleteFeedItem(ctx, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (api *API) PatchChannel(w http.ResponseWriter, r *http.Request) {
	var params models.UpdateFeedChannelParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := ValidateStruct(w, &params); err != nil {
		return
	}
	ctx := r.Context()
	if err := api.Queries.UpdateFeedChannel(ctx, params); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (api *API) PatchItem(w http.ResponseWriter, r *http.Request) {
	var params models.UpdateFeedItemParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := ValidateStruct(w, &params); err != nil {
		return
	}
	ctx := r.Context()
	if err := api.Queries.UpdateFeedItem(ctx, params); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (api *API) ListGroups(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	groups, err := api.Queries.ListGroup(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(groups)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func (api *API) AddGroup(w http.ResponseWriter, r *http.Request) {
	var params models.CreateGroupParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := ValidateStruct(w, &params); err != nil {
		return
	}
	ctx := r.Context()
	if err := api.Queries.CreateGroup(ctx, params); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (api *API) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	var params models.UpdateGroupParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := ValidateStruct(w, &params); err != nil {
		return
	}
	ctx := r.Context()
	if err := api.Queries.UpdateGroup(ctx, params); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)

}

func (api *API) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	if err := api.Queries.DeleteGroup(ctx, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (api *API) AddChannelToGroup(w http.ResponseWriter, r *http.Request) {
	var params models.AddChannelToGroupParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := ValidateStruct(w, &params); err != nil {
		return
	}
	ctx := r.Context()
	if err := api.Queries.AddChannelToGroup(ctx, params); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (api *API) RemoveChannelFromGroup(w http.ResponseWriter, r *http.Request) {
	var params models.RemoveChannelFromGroupParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := ValidateStruct(w, &params); err != nil {
		return
	}
	ctx := r.Context()
	if err := api.Queries.RemoveChannelFromGroup(ctx, params); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
