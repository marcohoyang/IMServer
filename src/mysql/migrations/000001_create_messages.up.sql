-- 创建消息表
CREATE TABLE IF NOT EXISTS messages (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    from_id BIGINT UNSIGNED NOT NULL COMMENT '发送者ID',
    to_id BIGINT UNSIGNED NOT NULL COMMENT '接收者ID（私聊为用户ID，群聊为群组ID）',
    type TINYINT NOT NULL COMMENT '消息类型：0-未知 1-私聊 2-群聊',
    content_type TINYINT NOT NULL COMMENT '内容类型：0-文本 1-图片 2-语音',
    content BLOB NOT NULL COMMENT '消息内容（二进制数据）',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (id),
    INDEX idx_from_id (from_id),
    INDEX idx_to_id (to_id),
    INDEX idx_type_to_id (type, to_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息表'; 