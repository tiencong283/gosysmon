import React from "react"
import "./About.css"
import Header from '../Header/Header'
import * as AuthService from "../Auth/AuthService";
import { Redirect } from "react-router-dom";

const title = "About - GoSysmon"

class About extends React.Component {
    componentDidMount() {
        document.title = title
    }

    render() {
        return (
            <div className="grid-container full">
                <div className="grid-x grid-margin-x main-container">
                    <Header />
                    <div className="cell auto content-wrapper">
                        <div className="inner-content-wrapper">
                            <div id="aboutme">
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        )
    }
}

export default About
