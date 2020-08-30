import { Button, FormGroup, InputGroup, Tab, Tabs } from "@blueprintjs/core";
import { Cell, Column, Table } from "@blueprintjs/table";
import React, { useState } from "react";

/*
Hosts:
    - Add new host
    - Host list
*/
export const Host: React.FC = () => {


    return (
        <div>
            <Tabs
                id="TabsExample"
                vertical={false}
            >
                <Tab id="rx" title="List" panel={<HostList />} />
                <Tab id="ng" title="Add" panel={<AddHost />} />
            </Tabs>
        </div>
    )
}

const HostList: React.FC = () => {
    const cellRenderer = (rowIndex: number) => <Cell>{`$${(rowIndex * 10).toFixed(2)}`}</Cell>;

    return (
        <Table numRows={10}>
            <Column name="Dollars" cellRenderer={cellRenderer} />
        </Table>
    )
}

const AddHost: React.FC = () => {
    const [text, setText] = useState("")

    return (
        <div>

            <FormGroup
                helperText=""
                label="IP"
                labelFor="text-input"
                labelInfo="(required)"
            >
                <InputGroup id="text-input" placeholder="127.0.0.1" />
            </FormGroup>

            <FormGroup
                helperText=""
                label="user name"
                labelFor="text-input"
                labelInfo="(required)"
            >
                <InputGroup id="text-input" placeholder="ubuntu" />
            </FormGroup>

            <FormGroup
                helperText=""
                label="Password"
                labelFor="text-input"
            // labelInfo="(required)"
            >
                <InputGroup id="text-input" placeholder="myPassword" />
            </FormGroup>

            <FormGroup label="Pem">
                <InputGroup placeholder="Pem file path" value={text} />
            </FormGroup>

            <Button text="Ping" icon="add" />
            <Button text="Add" icon="add" />
            <Button text="Deploy" icon="add" />

        </div>
    )
}