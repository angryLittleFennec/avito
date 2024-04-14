DROP TABLE IF EXISTS banners CASCADE;
DROP TABLE IF EXISTS banner_feature_tags CASCADE;

CREATE TABLE IF NOT EXISTS banners (
    id SERIAL PRIMARY KEY,
    content JSON NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN NOT NULL
);

CREATE TABLE IF NOT EXISTS banner_feature_tags (
    id SERIAL PRIMARY KEY,
    banner_id INTEGER NOT NULL,
    feature_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    CONSTRAINT fk_banner
        FOREIGN KEY (banner_id) 
        REFERENCES banners (id)
        ON DELETE CASCADE,
    CONSTRAINT idx_feature_tag
        UNIQUE (feature_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_banners_on_id ON banners (id);
CREATE INDEX IF NOT EXISTS idx_banner_feature_tags_on_banner_id ON banner_feature_tags (banner_id);
CREATE INDEX IF NOT EXISTS idx_banner_feature_tags_on_feature_id ON banner_feature_tags (feature_id);
CREATE INDEX IF NOT EXISTS idx_banner_feature_tags_on_tag_id ON banner_feature_tags (tag_id);

INSERT INTO banners (content, is_active) VALUES
('{"title": "Summer Sale", "description": "Up to 50% off!"}', TRUE),
('{"title": "Winter Sale", "description": "Holiday specials from 30% off!"}', TRUE);

INSERT INTO banner_feature_tags (banner_id, feature_id, tag_id) VALUES
(1, 1, 101),
(1, 1, 102),
(2, 2, 201),
(2, 2, 202);
