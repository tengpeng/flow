import { Card } from "@blueprintjs/core";
import React from "react";
import GridLayout, { WidthProvider } from "react-grid-layout";
import "react-grid-layout/css/styles.css";
import "react-resizable/css/styles.css";
import { FlowList, FlowRun, NewFlow } from "./flows";
import { AddHost, HostList } from "./hosts";
const GridLayoutWidth = WidthProvider(GridLayout)

/*
TODOs:
- divide page
- add title
- tune initial layout
- add top navbar
*/
export const Grid: React.FC = () => {
    const layout = [
        { i: 'AddHost', x: 0, y: 0, w: 2, h: 16 },
        { i: 'HostList', x: 2, y: 0, w: 2, h: 4 },
        { i: 'FlowRun', x: 4, y: 0, w: 6, h: 8 },
        { i: 'NewFlow', x: 10, y: 0, w: 2, h: 4 },
        { i: 'FlowList', x: 12, y: 0, w: 2, h: 4 },
    ];

    return (
        <GridLayoutWidth className="layout" layout={layout} cols={12} rowHeight={30} width={1200}>
            {/* <CardList /> */}
            <div key="AddHost">
                <Card elevation={3}>
                    <AddHost />
                </Card>
            </div>
            <div key="HostList">
                <Card elevation={3}>
                    <HostList />
                </Card>
            </div>
            <div key="NewFlow">
                <Card elevation={3}>
                    <NewFlow />
                </Card>
            </div>
            <div key="FlowRun">
                <Card elevation={3}>
                    <FlowRun />
                </Card>
            </div>
            <div key="FlowList">
                <Card elevation={3}>
                    <FlowList />
                </Card>
            </div>
        </GridLayoutWidth>
    )
}

//TODO: card container
// const CardList: React.FC = () => {
//     const components = [<AddHost />, <HostList />, <FlowList />, <FlowRun />, <NewFlow />]

//     return (
//         <div>
//             {components.map((component, idx) => {
//                 return (
//                     <Card elevation={3}>
//                         {component}
//                     </Card>
//                 )

//             })}
//         </div>
//     )
// }