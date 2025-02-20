package moov

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type CreateTransferArgs func(t *createTransferBuilder) callArg
type createTransferBuilder struct {
	idempotencyKey string
}

// Can be specified to overwrite a randomly generated one.
func WithTransferIdempotencyKey(key uuid.UUID) CreateTransferArgs {
	return func(t *createTransferBuilder) callArg {
		t.idempotencyKey = key.String()
		return IdempotencyKey(t.idempotencyKey)
	}
}

// CreateTransfer creates a new transfer
// https://docs.moov.io/api/index.html#tag/Transfers/operation/createTransfer
func (c Client) CreateTransfer(ctx context.Context, transfer CreateTransfer, options ...CreateTransferArgs) CreateTransferBuilder {
	builder := &createTransferBuilder{}
	callArgs := []callArg{
		AcceptJson(),
		JsonBody(transfer),
		WithTransferIdempotencyKey(uuid.New())(builder),
	}

	for _, opt := range options {
		callArgs = append(callArgs, opt(builder))
	}

	return CreateTransferBuilder{
		client:   c,
		ctx:      ctx,
		endpoint: Endpoint(http.MethodPost, pathTransfers),
		callArgs: callArgs,
	}
}

type CreateTransferBuilder struct {
	client   Client
	ctx      context.Context
	endpoint EndpointArg
	callArgs []callArg
}

// Started initiates the transfers request and doesn't wait beyond creating the transfer
func (r CreateTransferBuilder) Started() (*TransferStarted, error) {
	resp, err := r.client.CallHttp(r.ctx, r.endpoint, r.callArgs...)
	if err != nil {
		return nil, err
	}

	switch resp.Status() {
	case StatusCompleted:
		st, err := UnmarshalObjectResponse[TransferStarted](resp)
		return st, err
	case StatusStateConflict:
		return nil, errors.Join(ErrXIdempotencyKey, resp)
	default:
		return nil, resp
	}
}

// Starts a transfer request and waits for a response from the rail (e.g. authorized or declined) before returning the result.
// There are three possible return values (and only one will be not-nil):
// 1) A full transfer with rail-specific details as a result of waiting for the response from the rail.
// 2) A transfer that started but the request timed out waiting for a response from the rail.
// 3) An error attempting to create the transfer.
func (r CreateTransferBuilder) WaitForRailResponse() (*Transfer, *TransferStarted, error) {
	resp, err := r.client.CallHttp(r.ctx, r.endpoint, append(r.callArgs, WaitFor("rail-response"))...)
	if err != nil {
		return nil, nil, err
	}

	switch resp.Status() {
	case StatusCompleted:
		transfer, err := UnmarshalObjectResponse[Transfer](resp)
		return transfer, nil, err
	case StatusStarted:
		transferStarted, err := UnmarshalObjectResponse[TransferStarted](resp)
		return nil, transferStarted, err
	case StatusStateConflict:
		return nil, nil, errors.Join(ErrXIdempotencyKey, resp)
	default:
		return nil, nil, resp
	}
}

type ListTransferFilter callArg

func WithTransferAccountIDs(accountIDs []string) ListTransferFilter {
	return callBuilderFn(func(call *callBuilder) error {
		call.params["accountIDs"] = strings.Join(accountIDs, ",")
		return nil
	})
}

func WithTransferStatus(status string) ListTransferFilter {
	return callBuilderFn(func(call *callBuilder) error {
		call.params["status"] = status
		return nil
	})
}

func WithTransferStartDate(start time.Time) ListTransferFilter {
	return callBuilderFn(func(call *callBuilder) error {
		call.params["startDateTime"] = start.Format(time.RFC3339)
		return nil
	})
}

func WithTransferEndDate(end time.Time) ListTransferFilter {
	return callBuilderFn(func(call *callBuilder) error {
		call.params["endDateTime"] = end.Format(time.RFC3339)
		return nil
	})
}

func WithTransferGroup(groupID string) ListTransferFilter {
	return callBuilderFn(func(call *callBuilder) error {
		call.params["groupID"] = groupID
		return nil
	})
}

func WithTransferRefunded() ListTransferFilter {
	return callBuilderFn(func(call *callBuilder) error {
		call.params["refunded"] = "true"
		return nil
	})
}

func WithTransferDisputed() ListTransferFilter {
	return callBuilderFn(func(call *callBuilder) error {
		call.params["disputed"] = "true"
		return nil
	})
}

func WithTransferSkip(skip int) ListTransferFilter {
	return Skip(skip)
}

func WithTransferCount(count int) ListTransferFilter {
	return Count(count)
}

// ListTransfers lists all transfers
// https://docs.moov.io/api/index.html#tag/Transfers/operation/listTransfers
func (c Client) ListTransfers(ctx context.Context, filters ...ListTransferFilter) ([]Transfer, error) {
	resp, err := c.CallHttp(ctx,
		Endpoint(http.MethodGet, pathTransfers),
		prependArgs(filters, AcceptJson())...)
	if err != nil {
		return nil, err
	}

	return CompletedListOrError[Transfer](resp)
}

// GetTransfer retrieves a transfer
// https://docs.moov.io/api/index.html#tag/Transfers/operation/getTransfer
func (c Client) GetTransfer(ctx context.Context, transferID string) (*Transfer, error) {
	resp, err := c.CallHttp(ctx,
		Endpoint(http.MethodGet, pathTransfer, transferID),
		AcceptJson(),
	)
	if err != nil {
		return nil, err
	}

	return CompletedObjectOrError[Transfer](resp)
}

type TransferPatcher func(patch *patchTransfer)

func PatchTransferMetadata(metadata map[string]string) TransferPatcher {
	return func(patch *patchTransfer) {
		patch.Metadata = metadata
	}
}

// UpdateTransferMetaData updates the metadata for a transfer
// https://docs.moov.io/api/index.html#tag/Transfers/operation/patchTransfer
func (c Client) PatchTransfer(ctx context.Context, transferID string, patches ...TransferPatcher) (*Transfer, error) {
	patch := &patchTransfer{}
	for _, p := range patches {
		p(patch)
	}

	resp, err := c.CallHttp(ctx,
		Endpoint(http.MethodPatch, pathTransfer, transferID),
		AcceptJson(),
		JsonBody(patch))
	if err != nil {
		return nil, err
	}

	return CompletedObjectOrError[Transfer](resp)
}

type CreateRefundArgs callArg

func WithRefundWaitForRailResponse() CreateRefundArgs {
	return WaitFor("rail-response")
}

// Can be specified to overwrite a randomly generated one.
func WithRefundIdempotencyKey(key uuid.UUID) CreateRefundArgs {
	return IdempotencyKey(key.String())
}

// RefundTransfer refunds a transfer
// https://docs.moov.io/api/#tag/Transfers/operation/refundTransfer
func (c Client) RefundTransfer(ctx context.Context, transferID string, refund CreateRefund, options ...CreateRefundArgs) (*Refund, *RefundStarted, error) {
	args := prependArgs(options,
		AcceptJson(),
		WithRefundIdempotencyKey(uuid.New()),
		JsonBody(refund),
	)

	resp, err := c.CallHttp(ctx,
		Endpoint(http.MethodPost, pathRefunds, transferID),
		args...)
	if err != nil {
		return nil, nil, err
	}

	switch resp.Status() {
	case StatusCompleted:
		r, err := CompletedObjectOrError[Refund](resp)
		return r, nil, err
	case StatusStarted:
		r, err := CompletedObjectOrError[RefundStarted](resp)
		return nil, r, err
	default:
		return nil, nil, resp
	}
}

// ListRefunds lists all refunds for a transfer
// https://docs.moov.io/api/index.html#tag/Transfers/operation/getRefunds
func (c Client) ListRefunds(ctx context.Context, transferID string) ([]Refund, error) {
	resp, err := c.CallHttp(ctx,
		Endpoint(http.MethodGet, pathRefunds, transferID),
		AcceptJson())
	if err != nil {
		return nil, err
	}

	return CompletedListOrError[Refund](resp)
}

// GetRefund retrieves a refund for a transfer
// https://docs.moov.io/api/index.html#tag/Transfers/operation/getRefund
func (c Client) GetRefund(ctx context.Context, transferID string, refundID string) (*Refund, error) {
	resp, err := c.CallHttp(ctx, Endpoint(http.MethodGet, pathRefund, transferID, refundID), AcceptJson())
	if err != nil {
		return nil, err
	}

	return CompletedObjectOrError[Refund](resp)
}

type CreateReversalArgs callArg

// Can be specified to overwrite a randomly generated one.
func WithReversalsIdempotencyKey(key uuid.UUID) CreateReversalArgs {
	return IdempotencyKey(key.String())
}

// ReverseTransfer reverses a transfer
// https://docs.moov.io/api/index.html#tag/Transfers/operation/reverseTransfer
func (c Client) ReverseTransfer(ctx context.Context, transferID string, refund CreateReversal, options ...CreateReversalArgs) (*CreatedReversal, error) {
	args := prependArgs(options,
		AcceptJson(),
		WithReversalsIdempotencyKey(uuid.New()),
		JsonBody(refund),
	)

	resp, err := c.CallHttp(ctx, Endpoint(http.MethodPost, pathReversals, transferID), args...)
	if err != nil {
		return nil, err
	}

	return CompletedObjectOrError[CreatedReversal](resp)
}

// TransferOptions lists all transfer options between a source and destination
// https://docs.moov.io/api/#tag/Transfers/operation/createTransferOptions
func (c Client) TransferOptions(ctx context.Context, payload CreateTransferOptions) (*TransferOptions, error) {
	resp, err := c.CallHttp(ctx,
		Endpoint(http.MethodPost, pathTransferOptions),
		AcceptJson(),
		JsonBody(payload))
	if err != nil {
		return nil, err
	}

	return CompletedObjectOrError[TransferOptions](resp)
}
