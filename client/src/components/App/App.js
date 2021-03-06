import React from "react"
import "foundation-sites/dist/css/foundation.min.css"
import "./App.css"

import { BrowserRouter, Route, Switch } from "react-router-dom"
import Header from "../Header/Header"
import Home from "../Home/Home"
import AlertList from "../AlertList/AlertList"
import IOCList from "../IOCList/IOCList"
import HostList from "../HostList/HostList"
import ActivityLogList from "../ActivityLogList/ActivityLogList"
import About from "../About/About"
import ProtectedRoute from "../Auth/ProtectedRoute"
import NotifyLogin from "../Auth/NotifyLogin"
import Navbar from '../Navbar/Navbar'
import ProcessWrapper from "../Process/Process"

// main navigation
const navItems = [
    {
        name: 'Home',
        path: '/',
        icon: 'fi-home',
        component: Home,
    },
    {
        name: 'Alerts',
        path: '/alerts.html',
        icon: 'fi-alert',
        component: AlertList,
    },
    {
        name: 'IOCs',
        path: '/iocs.html',
        icon: 'fi-skull',
        component: IOCList,
    },
    {
        name: 'Hosts',
        path: '/hosts.html',
        icon: 'fi-social-windows',
        component: HostList,
    },
    {
        name: 'Activity Logs',
        path: '/activity-logs.html',
        icon: 'fi-clipboard-notes',
        component: ActivityLogList,
    },
    {
        name: 'About',
        path: '/about.html',
        icon: 'fi-info',
        component: About,
    },
]

class App extends React.Component {
    render() {
        return (
            <BrowserRouter>
                <Switch>
                    <Route path="/login" exact component={NotifyLogin} />
                    {
                        navItems.map((navItem, index) => (
                            <ProtectedRoute key={index} path={navItem.path} exact component={navItem.component} />
                        ))
                    }
                    <Route path="/process"><ProcessWrapper /></Route>
                </Switch>
            </BrowserRouter>
        )
    }
}

export default App
