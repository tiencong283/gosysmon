import React from "react"
import "./Header.css"
import { Link } from "react-router-dom"
import Home from "../Home/Home"
import AlertList from "../AlertList/AlertList"
import IOCList from "../IOCList/IOCList"
import HostList from "../HostList/HostList"
import ActivityLogList from "../ActivityLogList/ActivityLogList"
import About from "../About/About"


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

function active() {
    for (let i = 1; i < navItems.length; i++)
        if (window.location.href.includes(navItems[i].path) && navItems[i].path != "/")
	{
            return navItems[i].name
	}
    if (window.location.href.includes("/process"))
	return "Alerts"
    return "Home"
}


class Header extends React.Component {

    constructor(props) {
        super(props)
        this.state = {
            tab: active(),
        }
    }

    renderNavItems() {
        return navItems.map((navItem, idx) => {
            let active = this.state.tab === navItem.name ? "underline" : ""
            return (
                <li key={idx} className={active}>
                    <Link to={navItem.path}><span><i
                        className={`${navItem.icon} main-sidebar-icon`}/>{navItem.name}</span></Link>
                </li>
            )
        })
    }

    render() {
        return (
            <div className="cell medium-2 main-sidebar">
                <ul className="vertical menu">
                    {
                        this.renderNavItems()
                    }
                </ul>
            </div>
        )
    }
}

export default Header
