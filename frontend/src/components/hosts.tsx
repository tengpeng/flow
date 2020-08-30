import { Button, FormGroup, InputGroup, Tab, Tabs } from "@blueprintjs/core";
import React, { useState } from "react";
import { useTable } from 'react-table';

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
    // const cellRenderer = (rowIndex: number) => <Cell>{`$${(rowIndex * 10).toFixed(2)}`}</Cell>;
    // const cellRenderer = (rowIndex: number) => {
    //     return <Cell>{`$${(rowIndex * 10).toFixed(2)}`}</Cell>
    // }

    const [data, setData] = useState([{ ip: "123.456.789", actions: <Button text="Ping" icon="add" /> }] as any)

    const handleClick = () => {

    }

    const columns = [
        {
            Header: 'IP',
            accessor: 'ip',
        },
        {
            Header: 'Actions',
            accessor: 'action',
            Cell: () => (
                <div>
                    <button onClick={() => handleClick()}>New notebook</button>
                </div>
            )
        }
    ];



    const {
        getTableProps,
        getTableBodyProps,
        headerGroups,
        rows,
        prepareRow,
    } = useTable({
        columns,
        data,
    });

    // useEffect(() => {
    //     fetch(url)
    //         .then(response => response.json())
    //         .then(data => setData(data))
    // }, [])


    return (
        < table {...getTableProps()}>
            <thead>
                {headerGroups.map(headerGroup => (
                    <tr {...headerGroup.getHeaderGroupProps()}>
                        {headerGroup.headers.map(column => (
                            <th {...column.getHeaderProps()}>{column.render('Header')}</th>
                        ))}
                    </tr>
                ))}
            </thead>
            <tbody {...getTableBodyProps()}>
                {rows.map((row, i) => {
                    prepareRow(row)
                    return (
                        <tr {...row.getRowProps()}>
                            {row.cells.map(cell => {
                                return <td {...cell.getCellProps()}>{cell.render('Cell')}</td>
                            })}
                        </tr>
                    )
                })}
            </tbody>
        </table >

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

