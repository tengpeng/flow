import React from 'react';
import { RecoilRoot } from 'recoil';
import './App.css';
import { Grid } from './components/grid';

function App() {
    return (
        <RecoilRoot>
            <Grid />
        </RecoilRoot>
    );
}

export default App;
