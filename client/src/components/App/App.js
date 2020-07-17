import React from "react"
import "./App.css"
import Header from "../Header/Header"
import {BrowserRouter as Router, Route, Switch} from "react-router-dom"
import AlertList from "../AlertList/AlertList"
import IOCList from "../IOCList/IOCList"
import HostList from "../HostList/HostList"
import ActivityLogList from "../ActivityLogList/ActivityLogList"
import Home from "../Home/Home"

class App extends React.Component {
    render() {
        return (
            <Router>
                <div className="grid-container">
                    <div className="grid-x grid-margin-x main-container">
                        <Header/>
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
            </Router>
        )
    }
}

export default App