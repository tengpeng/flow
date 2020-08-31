import { Button, ControlGroup, FormGroup, InputGroup, MenuItem, Tab, Tabs } from "@blueprintjs/core";
import { ItemRenderer, Select } from "@blueprintjs/select";
import React, { useEffect, useState } from "react";
import { host } from "./hosts";

/*
flows:
    - list of flows
    - list of flow runs
    - list of task runs
    - new flow
*/

const baseURL = "http://127.0.0.1:9000"

export const Flow: React.FC = () => {


    return (
        <Tabs
            id="TabsExample"
            vertical={false}
        >
            <Tab id="1" title="List" panel={<FlowList />} />
            <Tab id="2" title="Add" panel={<FlowRun />} />
            <Tab id="3" title="New" panel={<NewFlow />} />
        </Tabs>
    )
}

interface flow {
    ID: string
    FlowName: string,
    // ip: string, TODO: hostID
    Schedule: string,
}

const FlowList: React.FC = () => {
    const url = baseURL + "/flows"
    const [flows, setFlows] = useState<flow[]>([])

    useEffect(() => {
        async function fetchData() {
            const res = await fetch(url)
            res
                .json()
                .then(res => setFlows(res))
                .catch(err => alert(err))
        }
        fetchData()

    }, [url])

    //TODO: start flows
    const setRows = () => {
        return flows.map((flow, index) =>
            <tbody key={index}>
                <tr>
                    <td>{flow.FlowName}</td>
                    <td>{flow.Schedule}</td>
                    <td>  <Button text="start" onClick={() => handleStart(flow.ID)} /></td>
                </tr>
            </tbody>
        )
    }

    const handleStart = (ID: string) => {
        const url = baseURL + "/flows/" + ID
        fetch(url,
            {
                method: 'post',
            })
            .then(response => response.json())
            .then(result => {
                if (result.error !== undefined) {
                    alert(result.error)
                } else {
                    alert(result.message)
                }
            })
            .catch(error => {
                alert(error)
            });
    }

    return (
        <div>
            <table className="bp3-html-table bp3-html-table-bordered bp3-html-table-striped">
                <thead>
                    <tr>
                        <th>Name</th>
                        <th>Schedule</th>
                        <th>Actions</th>
                    </tr>
                </thead>
                {setRows()}
            </table>
        </div>
    )
}

interface run {
    FlowName: string,
    HostID: string,
    Status: string
}

//TODO: how to view task run & notebook
const FlowRun: React.FC = () => {
    const url = baseURL + "/runs"
    const [runs, setRuns] = useState<run[]>([])

    useEffect(() => {
        async function fetchData() {
            const res = await fetch(url)
            res
                .json()
                .then(res => setRuns(res))
                .catch(err => alert(err))
        }
        fetchData()

    }, [url])

    //TODO: start flows
    const setRows = () => {
        return runs.map((run, index) =>
            <tbody key={index}>
                <tr>
                    <td>{run.FlowName}</td>
                    <td>{run.HostID}</td>
                    <td>{run.Status}</td>
                </tr>
            </tbody>
        )
    }

    return (
        <div>
            <table className="bp3-html-table bp3-html-table-bordered bp3-html-table-striped">
                <thead>
                    <tr>
                        <th>Name</th>
                        <th>Host</th>
                        <th>Status</th>
                    </tr>
                </thead>
                {setRows()}
            </table>
        </div>
    )
}

const HostSelect = Select.ofType<host>();

interface task {
    Name: string
    Path: string
    Next: string
}

const NewFlow: React.FC = () => {
    const url = baseURL + "/hosts"

    const [name, setName] = useState("")
    const [schedule, setSchedule] = useState("")
    const [hosts, setHosts] = useState<host[]>([])

    useEffect(() => {
        async function fetchData() {
            const res = await fetch(url)
            res
                .json()
                .then(res => setHosts(res))
                .catch(err => alert(err))
        }
        fetchData()

    }, [url])

    const handleValueChange = () => {

    }

    const renderHost: ItemRenderer<host> = (host, { handleClick, modifiers, query }) => {
        if (!modifiers.matchesPredicate) {
            return null;
        }
        console.log("host: ", host)
        return (
            <MenuItem
                active={modifiers.active}
                disabled={modifiers.disabled}
                key={host.IP}
                onClick={handleClick}
                text={host.IP}
            />
        );
    };

    return (
        <div className="add-host">

            <ControlGroup fill={false} vertical={true}>

                <FormGroup
                    helperText=""
                    label="Name"
                    labelFor="text-input"
                    labelInfo="(required)"
                >
                    <InputGroup id="text-input"
                        placeholder="wf1"
                        value={name}
                        onChange={(e: React.ChangeEvent<HTMLInputElement>) => setName(e.target.value)} />
                </FormGroup>

                <HostSelect
                    itemRenderer={renderHost}
                    items={hosts}
                    onItemSelect={handleValueChange}
                >
                    <Button
                        icon="film"
                        rightIcon="caret-down"
                        text={"(No selection)"}
                    />
                </HostSelect>

                <FormGroup
                    helperText=""
                    label="Schedule"
                    labelFor="text-input"
                    labelInfo="(required)"
                >
                    <InputGroup id="text-input"
                        placeholder="*/5 * * * *"
                        value={schedule}
                        onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSchedule(e.target.value)} />
                </FormGroup>
            </ControlGroup>
            <Tasks />
        </div >
    )
}

const Tasks: React.FC = () => {
    const [tasks, setTasks] = useState<task[]>([{ Name: "", Path: "", Next: "" }]);

    const handleAddClick = () => {
        setTasks(tasks.concat({} as task));
    };


    const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>, idx: number) => {
        const { name, value } = e.target;
        let copy: task[] = Object.assign([], tasks);  // creating copy of state variable jasper
        (copy[idx] as any)[name] = value
        setTasks(copy);
    };


    const handleRemoveClick = (index: string) => {
        // const list = [...tasks];
        // list.splice(index, 1);
        // setTasks(list);
    };

    // handle click event of the Add button

    return (
        <div>
            {tasks.map((task, idx) => {
                return (
                    <div>
                        <ControlGroup fill={false} vertical={true}>

                            <FormGroup
                                helperText=""
                                label="Name"
                                labelFor="text-input"
                                labelInfo="(required)"
                            >
                                <InputGroup id="text-input"
                                    placeholder="nb1"
                                    name={"Name"}
                                    value={task.Name}
                                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleInputChange(e, idx)} />
                            </FormGroup>

                            <FormGroup
                                helperText=""
                                label="Path"
                                labelFor="text-input"
                                labelInfo="(required)"
                            >
                                <InputGroup id="text-input"
                                    placeholder="/Users/Downloads/nb1.ipynb"
                                    name={"Path"}
                                    value={task.Path}
                                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleInputChange(e, idx)} />
                            </FormGroup>

                            <FormGroup
                                helperText=""
                                label="Next"
                                labelFor="text-input"
                                labelInfo="(required)"
                            >
                                <InputGroup id="text-input"
                                    placeholder="nb3"
                                    name={"Next"}
                                    value={task.Next}
                                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleInputChange(e, idx)} />
                            </FormGroup>

                        </ControlGroup>

                        <Button text="Add" onClick={handleAddClick} />
                    </div>
                );
            })}

        </div>
    );
}

