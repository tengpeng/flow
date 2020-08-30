import React from 'react';
import SplitPane from 'react-split-pane';
import './App.css';
import { Host } from './components/hosts';
import { Sidebar } from './components/sidebar';

function App() {
    return (
        <SplitPane split="vertical" defaultSize="5%">
            <div className="Panel-1" >
                <Sidebar />
            </div>

            <SplitPane split="vertical" defaultSize="15%" >
                <div className="Panel-2" >
                </div>

                <div className="Panel-3">
                    <Host />
                </div>

            </SplitPane>
        </SplitPane>
    );
}

export default App;
