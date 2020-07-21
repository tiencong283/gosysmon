import React from "react"
import "./Home.css"

const title = "Home - GoSysmon"

class Home extends React.Component {
    componentDidMount() {
        document.title = title
    }

    render() {
        return (
            <h1>Home</h1>
        )
    }
}

export default Home