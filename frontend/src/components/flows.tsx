import { Button, Tab, Tabs } from "@blueprintjs/core"
import React, { useEffect, useState } from "react"

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
            <Tab id="rx" title="List" panel={<FlowList />} />
            <Tab id="ng" title="Add" panel={<FlowRun />} />
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

const NewFlow: React.FC = () => {

    return (
        <div>Flow</div>
    )
}