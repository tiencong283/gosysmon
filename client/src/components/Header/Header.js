import React from "react"
import "./Header.css"
import {Link} from "react-router-dom"

class Header extends React.Component {
    render() {
        return (
                <div className="cell medium-2 main-sidebar">
                    <ul className="vertical menu">
                        <li><Link to="/"><span><i className="fi-home main-sidebar-icon"></i>Home</span></Link></li>
                        <li><Link to="/alerts.html"><span><i className="fi-alert main-sidebar-icon"></i>Alerts</span></Link></li>
                        <li><Link to="/iocs.html"><span><i className="fi-skull main-sidebar-icon"></i>IOCs</span></Link></li>
                        <li><Link to="/hosts.html"><span><i className="fi-social-windows main-sidebar-icon"></i>Hosts</span></Link></li>
                        <li><Link to="/activity-logs.html"><span><i className="fi-clipboard-notes main-sidebar-icon"></i>Activity Logs</span></Link></li>
                    </ul>
                </div>
        )
    }
}

export default Header