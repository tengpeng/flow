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
        { i: 'c0', x: 0, y: 0, w: 2, h: 16 },
        { i: 'c1', x: 2, y: 0, w: 2, h: 16 },
        { i: 'c2', x: 4, y: 0, w: 2, h: 16 },
        { i: 'c3', x: 6, y: 0, w: 4, h: 16 },
        { i: 'c4', x: 12, y: 0, w: 2, h: 16 },
    ];

    const components = [<AddHost />, <HostList />, <FlowList />, <FlowRun />, <NewFlow />]

    components.map((component, idx) => console.log(idx))

    return (
        <GridLayoutWidth className="layout" layout={layout} cols={12} rowHeight={30}>
            {components.map((component, idx) => <div key={'c' + idx}><CardContaienr key={'c' + idx} id={'c' + idx} component={component} /></div>)}
        </GridLayoutWidth>
    )
}

// height: 100 %;
// overflow: auto;

interface Props {
    id: string,
    component: any //TODO
    children?: any
}
//TODO: card container
const CardContaienr: React.FC<Props> = ({ children, id, component }) => {
    return (
        < div key={id} className="card-container">
            <Card elevation={3} >
                {component}
                {children}
            </Card>
        </div >
    )
}
