package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"interview/internal/model"
	"interview/internal/service"
)

// WalletHandler 处理钱包操作的HTTP请求
type WalletHandler struct {
	service *service.WalletService
}

// NewWalletHandler 创建一个新的钱包处理器
func NewWalletHandler(service *service.WalletService) *WalletHandler {
	return &WalletHandler{
		service: service,
	}
}

// CreateWallet 创建新钱包
// @Summary 创建新钱包
// @Description 创建一个新的钱包，自动生成唯一ID
// @Tags 钱包管理
// @Accept json
// @Produce json
// @Success 201 {object} model.CreateWalletResponse "创建成功"
// @Failure 500 {object} model.ErrorResponse "服务器内部错误"
// @Router /wallets [post]
func (h *WalletHandler) CreateWalletHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, "不允许的请求方法", http.StatusMethodNotAllowed)
		return
	}

	wallet, err := h.service.CreateWallet()
	if err != nil {
		h.writeError(w, "创建钱包失败", http.StatusInternalServerError)
		return
	}

	response := model.CreateWalletResponse{
		ID: wallet.ID,
	}

	h.writeJSON(w, response, http.StatusCreated)
}

// GetWallet 获取钱包信息
// @Summary 获取钱包信息
// @Description 根据钱包ID获取钱包的详细信息，包括余额
// @Tags 钱包管理
// @Accept json
// @Produce json
// @Param id path string true "钱包ID" Format(uuid)
// @Success 200 {object} model.WalletResponse "获取成功"
// @Failure 400 {object} model.ErrorResponse "无效的请求参数"
// @Failure 404 {object} model.ErrorResponse "钱包不存在"
// @Failure 500 {object} model.ErrorResponse "服务器内部错误"
// @Router /wallets/{id} [get]
func (h *WalletHandler) GetWalletHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, "不允许的请求方法", http.StatusMethodNotAllowed)
		return
	}

	// 从路径中提取钱包ID
	id := r.URL.Path[len("/wallets/"):]
	if id == "" || id == "/wallets" {
		h.writeError(w, "需要提供钱包ID", http.StatusBadRequest)
		return
	}

	wallet, err := h.service.GetWallet(id)
	if err != nil {
		if errors.Is(err, service.ErrWalletNotFound) {
			h.writeError(w, "钱包不存在", http.StatusNotFound)
			return
		}
		h.writeError(w, "获取钱包失败", http.StatusInternalServerError)
		return
	}

	response := model.WalletResponse{
		ID:      wallet.ID,
		Balance: wallet.Balance,
	}

	h.writeJSON(w, response, http.StatusOK)
}

// Transfer 钱包间转账
// @Summary 钱包间转账
// @Description 从一个钱包转账到另一个钱包，金额必须大于0且源钱包余额充足
// @Tags 钱包管理
// @Accept json
// @Produce json
// @Param request body model.TransferRequest true "转账请求"
// @Success 200 {object} model.TransferResponse "转账成功"
// @Failure 400 {object} model.ErrorResponse "无效的请求参数"
// @Failure 404 {object} model.ErrorResponse "钱包不存在"
// @Failure 500 {object} model.ErrorResponse "服务器内部错误"
// @Router /wallets/transfer [post]
func (h *WalletHandler) TransferHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, "不允许的请求方法", http.StatusMethodNotAllowed)
		return
	}

	var req model.TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "无效的请求体", http.StatusBadRequest)
		return
	}

	// 验证请求
	if req.FromWalletID == "" || req.ToWalletID == "" {
		h.writeError(w, "需要提供 from_wallet_id 和 to_wallet_id", http.StatusBadRequest)
		return
	}

	response, err := h.service.Transfer(req.FromWalletID, req.ToWalletID, req.Amount)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrWalletNotFound):
			h.writeError(w, "一个或两个钱包不存在", http.StatusNotFound)
		case errors.Is(err, service.ErrInsufficientBalance):
			h.writeError(w, "源钱包余额不足", http.StatusBadRequest)
		case errors.Is(err, service.ErrInvalidAmount):
			h.writeError(w, "金额必须为正数", http.StatusBadRequest)
		case errors.Is(err, service.ErrSameWallet):
			h.writeError(w, "不能转账到同一钱包", http.StatusBadRequest)
		default:
			h.writeError(w, "转账失败", http.StatusInternalServerError)
		}
		return
	}

	h.writeJSON(w, response, http.StatusOK)
}

// writeJSON 写入带有给定状态码的JSON响应
func (h *WalletHandler) writeJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError 写入带有给定状态码的错误响应
func (h *WalletHandler) writeError(w http.ResponseWriter, message string, status int) {
	h.writeJSON(w, model.ErrorResponse{Error: message}, status)
}
