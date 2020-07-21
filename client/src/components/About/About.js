import React from "react"
import "./About.css"

const title = "About - GoSysmon"

class About extends React.Component {
    componentDidMount() {
        document.title = title
    }

    render() {
        return (
            <h1>About</h1>
        )
    }
}

export default About