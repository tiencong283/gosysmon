import React from "react"
import "./About.css"

const title = "About - GoSysmon"

class About extends React.Component {
    componentDidMount() {
        document.title = title
    }

    render() {
        return (
            <div className="inner-content-wrapper">
                <div id="aboutme">
                </div>
            </div>
        )
    }
}

export default About