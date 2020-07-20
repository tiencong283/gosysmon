import React from "react"
import "./Header.css"
import {Link} from "react-router-dom"

class Header extends React.Component {
    render() {
        return (
            <div className="cell medium-2 main-sidebar">
                <ul className="vertical menu">
                    {
                        this.props.navItems.map(function (navItem, index) {
                            return (
                                <li key={index}>
                                    <Link to={navItem.path}><span><i className={`${navItem.icon} main-sidebar-icon`}></i>{navItem.name}</span></Link>
                                </li>
                            )
                        })
                    }
                </ul>
            </div>
        )
    }
}

export default Header