import React from "react"
import "./Header.css"
import {Link} from "react-router-dom"

class Header extends React.Component {

    constructor(props) {
        super(props)
        this.state = {
            activeIdx: 0
        }
    }

    handleChangeNavItem(idx) {
        this.setState({
            activeIdx: idx
        })
    }

    renderNavItems() {
        return this.props.navItems.map((navItem, idx) => {
            let active = this.state.activeIdx === idx ? "underline" : ""
            return (
                <li key={idx} className={active} onClick={this.handleChangeNavItem.bind(this, idx)}>
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