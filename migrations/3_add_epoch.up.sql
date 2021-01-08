-- Copyright 2019 Cray Inc. All Rights Reserved.
--
-- Except as permitted by contract or express written permission of Cray Inc.,
-- no part of this work or its content may be modified, used, reproduced or
-- disclosed in any form. Modifications made without express permission of
-- Cray Inc. may damage the system the software is installed within, may
-- disqualify the user from receiving support from Cray Inc. under support or
-- maintenance contracts, or require additional support services outside the
-- scope of those contracts to repair the software or system.

ALTER TABLE components
    ADD last_updated_version bigint NOT NULL
    DEFAULT 1;

ALTER TABLE components
    ADD CONSTRAINT version_constraint FOREIGN KEY (last_updated_version)
    REFERENCES version_history(version);
    
ALTER TABLE network
    ADD last_updated_version bigint NOT NULL
    DEFAULT 1;

ALTER TABLE network
    ADD CONSTRAINT version_constraint FOREIGN KEY (last_updated_version)
    REFERENCES version_history(version);