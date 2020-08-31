import { Divider, Tab, TabId, Tabs } from "@blueprintjs/core";
import React, { useState } from "react";
import { Flow } from "./flows";
import { Host } from "./hosts";

export const Sidebar: React.FC = () => {
    const [id, setID] = useState<TabId>("flow")

    return (
        <Tabs vertical={true} selectedTabId={id} onChange={(newTabId: TabId) => setID(newTabId)}>

            <Divider />

            <Tab title="host" id="host" panel={<Host />} />
            <Tab title="flow" id="flow" panel={<Flow />} />
        </Tabs>
    )
}