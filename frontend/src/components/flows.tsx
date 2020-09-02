import { Button, ControlGroup, Dialog, Divider, FormGroup, InputGroup, MenuItem } from "@blueprintjs/core";
import { ItemRenderer, Select } from "@blueprintjs/select";
import React, { useEffect, useState } from "react";
import { atom, selector, useRecoilState, useRecoilValue } from "recoil";
import { host } from "./hosts";


/*
flows:
    - list of flows
    - list of flow runs
    - list of task runs
    - new flow
*/

const baseURL = "http://127.0.0.1:9000"

const flowRefresh = atom({
    key: 'flowRefresh', // unique ID (with respect to other atoms/selectors)
    default: 0, // default value (aka initial value)
});

interface flow {
    ID: string
    FlowName: string,
    // ip: string, TODO: hostID
    Schedule: string,
    Tasks: task[],
}

//TODO: add actions to stop/delete/show
export const FlowList: React.FC = () => {
    const url = baseURL + "/flows"
    const [flows, setFlows] = useState<flow[]>([])
    const flowRefreshTarget = selector({
        key: 'flowRefreshTarget', // unique ID (with respect to other atoms/selectors)
        get: ({ get }) => {
            return get(flowRefresh)
        },
    });
    const count = useRecoilValue(flowRefreshTarget);

    useEffect(() => {
        async function fetchData() {
            const res = await fetch(url)
            res
                .json()
                .then(res => setFlows(res))
                .catch(err => alert(err))
        }
        fetchData()

    }, [url, count])

    //TODO: start flows
    const setRows = () => {
        return flows.map((flow, index) =>
            <tbody key={index}>
                <tr>
                    <td>{flow.FlowName}</td>

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
        <>
            <h4>Flows </h4>
            <Divider />
            <table className="bp3-html-table bp3-html-table-bordered bp3-html-table-striped">
                <thead>
                    <tr>
                        <th>Name</th>

                        <th>Actions</th>
                    </tr>
                </thead>
                {setRows()}
            </table>
        </>
    )
}

interface run {
    FlowName: string,
    UpdatedAt: string,
    HostID: string,
    Status: string
    TaskRuns: TaskRun[]
}

//TODO: popover?
interface TaskRun {
    Name: string,
    Status: string,
    Notebook: string
}

interface taskRunProps {
    tasks: TaskRun[]
}

const taskRunStyle = {
    'alignSelf': 'center',
}

const TaskRun: React.FC<taskRunProps> = ({ tasks }) => {
    const [isOpen, setIsOpen] = useState(false)

    const handleClick = (Notebook: string) => {
        var w = window.open();
        w!.document.open();
        w!.document.write(Notebook);
        w!.document.close();
    }

    const setRows = () => {
        console.log(tasks)
        return tasks.map((task, index) =>
            <tbody key={index}>
                <tr>
                    <td>{task.Name}</td>
                    <td>{task.Status}</td>
                    <td><Button text="Open" onClick={() => handleClick(task.Notebook)} /></td>
                </tr>
            </tbody>)
    }


    return (
        <>
            <Button text="Show" onClick={() => setIsOpen(true)}></Button>

            <Dialog
                title="Tasks"
                isOpen={isOpen}
                onClose={() => setIsOpen(false)}
            >

                <table className="bp3-html-table bp3-html-table-bordered bp3-html-table-striped"
                    style={taskRunStyle}>
                    <thead>
                        <tr>
                            <th>Name</th>
                            <th>Status</th>
                            <th>Notebook</th>
                        </tr>
                    </thead>
                    {setRows()}
                </table>

            </Dialog>
        </>
    )
}

//TODO: trim time
export const FlowRun: React.FC = () => {
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
                    <td>{run.UpdatedAt}</td>
                    <td>{run.Status}</td>
                    <td><TaskRun tasks={run.TaskRuns} /></td>
                </tr>
            </tbody>
        )
    }

    return (
        <>
            <h4>Runs </h4>
            <Divider />
            <table className="bp3-html-table bp3-html-table-bordered bp3-html-table-striped">
                <thead>
                    <tr>
                        <th>Name</th>
                        <th>Time</th>
                        <th>Status</th>
                        <th>Task</th>
                    </tr>
                </thead>
                {setRows()}
            </table>
        </>
    )
}

const HostSelect = Select.ofType<host>();

interface task {
    Name: string
    Path: string
    Next: string
}

export const NewFlow: React.FC = () => {
    const url = baseURL + "/hosts"

    const [name, setName] = useState("")
    const [schedule, setSchedule] = useState("")
    const [hosts, setHosts] = useState<host[]>([])
    const [host, setHost] = useState<host>()
    const [tasks, setTasks] = useState<task[]>()
    const [count, setCount] = useRecoilState(flowRefresh);


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

    //TODO:
    const handleValueChange = (item: host) => {
        setHost(item)
    }

    const onTaskChange = (tasks: task[]) => {
        setTasks(tasks)
    }

    const handleSubmit = () => {
        const flow = {
            Name: name,
            Host: host,
            Schedule: schedule,
            Tasks: tasks,
        }

        console.log(flow)
        const url = baseURL + "/flows"
        fetch(url,
            {
                method: 'post',
                body: JSON.stringify(flow)
            })
            .then(response => response.json())
            .then(result => {
                if (result.error !== undefined) {
                    alert(result.error)
                } else {
                    alert(result.message)
                    setCount(count + 1)
                }
            })
            .catch(error => {
                alert(error)
            });
    }

    const renderHost: ItemRenderer<host> = (host, { handleClick, modifiers, query }) => {
        if (!modifiers.matchesPredicate) {
            return null;
        }
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
        <>
            <h4>Flows: New </h4>
            <Divider />

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
                    <text>Host</text>
                    <Button
                        rightIcon="caret-down"
                        text={"(Select host)"}
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
                <Divider />
            </ControlGroup>
            <Tasks onTaskChange={onTaskChange} />

            <Divider />
            <Button text="Submit" onClick={handleSubmit} />
        </ >
    )
}

interface Props {
    onTaskChange: (tasks: task[]) => void;
}

const Tasks: React.FC<Props> = ({ onTaskChange }) => {
    const [tasks, setTasks] = useState<task[]>([{ Name: "", Path: "", Next: "" }]);

    const handleAddClick = () => {
        setTasks(tasks.concat({} as task));
    };


    const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>, idx: number) => {
        const { name, value } = e.target;
        const copy = [...tasks];
        (copy[idx] as any)[name] = value
        setTasks(copy);
    };


    const handleRemoveClick = (index: number) => {
        const list = [...tasks];
        list.splice(index, 1);
        setTasks(list);
    };

    useEffect(() => {
        onTaskChange(tasks)
    }, [tasks, onTaskChange])

    // handle click event of the Add button

    return (
        <div>

            {tasks.map((task, idx) => {
                return (
                    <div key={idx}>
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
                            >
                                <InputGroup id="text-input"
                                    placeholder="nb3"
                                    name={"Next"}
                                    value={task.Next}
                                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleInputChange(e, idx)} />
                            </FormGroup>

                        </ControlGroup>
                        <Button text="Remove" onClick={() => handleRemoveClick(idx)} />
                        <Button text="Add" onClick={handleAddClick} />
                    </div>
                );
            })}

        </div>
    );
}

