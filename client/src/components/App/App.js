import React from "react"
import "./App.css"
import Header from "../Header/Header"
import {BrowserRouter, Route, Switch} from "react-router-dom"
import AlertList from "../AlertList/AlertList"
import IOCList from "../IOCList/IOCList"
import HostList from "../HostList/HostList"
import ActivityLogList from "../ActivityLogList/ActivityLogList"
import Home from "../Home/Home"

const navItems = [
    {
        name: 'Home',
        path: '/',
        icon: 'fi-home',
    },
    {
        name: 'Alerts',
        path: '/alerts.html',
        icon: 'fi-alert',
    },
    {
        name: 'IOCs',
        path: '/iocs.html',
        icon: 'fi-skull',
    },
    {
        name: 'Hosts',
        path: '/hosts.html',
        icon: 'fi-social-windows',
    },
    {
        name: 'Activity Logs',
        path: '/activity-logs.html',
        icon: 'fi-clipboard-notes',
    },
]

class App extends React.Component {
    render() {
        return (
            <BrowserRouter>
                <div className="grid-container">
                    <div className="grid-x grid-margin-x main-container">
                        <Header navItems={navItems}/>
                        <div className="cell auto main-content">
                            <Switch>
                                <Route path="/alerts.html"><AlertList/></Route>
                                <Route path="/iocs.html"><IOCList/></Route>
                                <Route path="/hosts.html"><HostList/></Route>
                                <Route path="/activity-logs.html"><ActivityLogList/></Route>
                                <Route path="/"><Home/></Route>
                            </Switch>
                        </div>
                    </div>
                </div>
            </BrowserRouter>
        )
    }
}

export default App