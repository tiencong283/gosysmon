import React from "react"
import "./Process.css"
import {useLocation} from "react-router-dom";

// A custom hook that builds on useLocation to parse
// the query string for you.
function useQuery() {
    return new URLSearchParams(useLocation().search)
}

export default function ProcessWrapper() {
    let query = useQuery()
    return (
        <Process providerGuid={query.get("ProviderGUID")} processGuid={query.get("ProcessGuid")}/>
    )
}

class Process extends React.Component {
    render() {
        return (
            <div>
                <h3>{this.props.providerGuid}</h3>
                <h3>{this.props.processGuid}</h3>
            </div>
        )
    }
}
