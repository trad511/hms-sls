-- Copyright 2019 Cray Inc. All Rights Reserved.
--
-- Except as permitted by contract or express written permission of Cray Inc.,
-- no part of this work or its content may be modified, used, reproduced or
-- disclosed in any form. Modifications made without express permission of
-- Cray Inc. may damage the system the software is installed within, may
-- disqualify the user from receiving support from Cray Inc. under support or
-- maintenance contracts, or require additional support services outside the
-- scope of those contracts to repair the software or system.

-------------
-- Components
-------------

CREATE TABLE components (
    xname            VARCHAR NOT NULL
        CONSTRAINT components_xname_pk
            PRIMARY KEY,
    parent           VARCHAR NOT NULL,
    comp_type        VARCHAR NOT NULL,
    comp_class       VARCHAR NOT NULL,
    extra_properties JSONB
);

CREATE UNIQUE INDEX components_xname_uindex
    ON components(xname);

CREATE INDEX components_parent_index
    ON components(parent);

CREATE INDEX components_comp_class_index
    ON components(comp_class);

CREATE INDEX components_comp_type_index
    ON components(comp_type);


-------------
-- Network
-------------

CREATE TABLE network (
    name      VARCHAR NOT NULL
        CONSTRAINT network_name_pk
            PRIMARY KEY,
    full_name VARCHAR NOT NULL,
    ip_ranges INET[]  NOT NULL,
    type      VARCHAR NOT NULL
);

CREATE INDEX network_full_name_index
    ON network(full_name);

CREATE INDEX network_ip_ranges_index
    ON network(ip_ranges);

CREATE INDEX network_type_index
    ON network(type);


-------------
-- Version
-------------

CREATE TABLE version_history (
    version        BIGSERIAL   NOT NULL
        CONSTRAINT version_history_version_pk
            PRIMARY KEY,
    timestamp      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_entity VARCHAR
);

CREATE INDEX version_history_timestamp_index
    ON version_history(timestamp);

CREATE INDEX version_history_updated_entity_index
    ON version_history(updated_entity);

/*
 * MIT License
 *
 * (C) Copyright [2019-2021] Hewlett Packard Enterprise Development LP
 *
 * Permission is hereby granted, free of charge, to any person obtaining a
 * copy of this software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation
 * the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the
 * Software is furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included
 * in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 * THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 * OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 * ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 * OTHER DEALINGS IN THE SOFTWARE.
 */

-- Insert a first version to prevent immediate queries on the version table from failing.

INSERT INTO
    version_history(updated_entity)
VALUES
    ('base')
