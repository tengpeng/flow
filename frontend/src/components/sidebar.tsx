import { Tab, Tabs } from "@blueprintjs/core";
import React from "react";

export const Sidebar: React.FC = () => {

    return (
        <Tabs vertical={true}>
            <Tab id="Remote" title="Remote" />
            <Tab id="Flow" title="Flow" />
        </Tabs>
    )
}