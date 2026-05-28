package clients

import (
	"context"
	"strings"

	"github.com/Zany2/browserflow/backend/api/clients/v1"
	"github.com/Zany2/browserflow/backend/internal/dao"
	"github.com/Zany2/browserflow/backend/internal/model/entity"
	websockets "github.com/Zany2/browserflow/backend/utility/websocket"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/util/gconv"
)

// ClientDetail returns one registered client node. 返回单个客户端节点
func (c *ControllerV1) ClientDetail(ctx context.Context, req *v1.ClientDetailReq) (res *v1.ClientDetailRes, err error) {
	record, err := queryClientRecord(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return &v1.ClientDetailRes{}, nil
	}

	client, err := clientRecordToEntity(record)
	if err != nil {
		return nil, err
	}
	return &v1.ClientDetailRes{Client: client}, nil
}

func queryClientRecord(ctx context.Context, id string) (gdb.Record, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, nil
	}

	columns := dao.Clients.Columns()
	record, err := dao.Clients.Ctx(ctx).Where(columns.NodeId, id).One()
	if err != nil || !record.IsEmpty() {
		return record, err
	}

	record, err = dao.Clients.Ctx(ctx).Where(columns.ClientIp, id).One()
	if err != nil || !record.IsEmpty() {
		return record, err
	}

	primaryID := gconv.Int64(id)
	if primaryID <= 0 {
		return record, nil
	}
	return dao.Clients.Ctx(ctx).WherePri(primaryID).One()
}

func scopedClientModel(ctx context.Context, record gdb.Record) *gdb.Model {
	columns := dao.Clients.Columns()
	model := dao.Clients.Ctx(ctx)
	if record.IsEmpty() {
		return model.Where(columns.Id, 0)
	}
	if nodeID := clientNodeIDFromRecord(record); nodeID != "" {
		if clientIP := clientIPFromRecord(record); clientIP != "" {
			return model.Where(columns.ClientIp, clientIP).Where(columns.NodeId, nodeID)
		}
		return model.Where(columns.NodeId, nodeID)
	}
	if primaryID := gconv.Int64(record[columns.Id]); primaryID > 0 {
		return model.WherePri(primaryID)
	}
	return model.Where(columns.ClientIp, clientIPFromRecord(record))
}

func clientIPFromRecord(record gdb.Record) string {
	if record.IsEmpty() {
		return ""
	}
	return strings.TrimSpace(gconv.String(record[dao.Clients.Columns().ClientIp]))
}

func clientPrimaryIDFromRecord(record gdb.Record) string {
	if record.IsEmpty() {
		return ""
	}
	return strings.TrimSpace(gconv.String(record[dao.Clients.Columns().Id]))
}

func clientNodeIDFromRecord(record gdb.Record) string {
	if record.IsEmpty() {
		return ""
	}
	return strings.TrimSpace(gconv.String(record[dao.Clients.Columns().NodeId]))
}

func clientConnectionIDFromRecord(record gdb.Record) string {
	return websockets.NodeConnectionID(clientIPFromRecord(record), clientNodeIDFromRecord(record))
}

func clientRecordToEntity(record gdb.Record) (*entity.Clients, error) {
	if record.IsEmpty() {
		return nil, nil
	}
	client := &entity.Clients{}
	if err := gconv.Struct(record, client); err != nil {
		return nil, err
	}
	return client, nil
}
