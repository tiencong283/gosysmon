import React from "react"
import "./Process.css"
import {useLocation} from "react-router-dom"
import $ from "jquery"

const endpoint = "/api/process"
const procStates = ["Running", "Stopped"]

// A custom hook that builds on useLocation to parse
// the query string for you.
function useQuery() {
    return new URLSearchParams(useLocation().search)
}

export default function ProcessWrapper() {
    let query = useQuery()
    return (
        <Process providerGuid={query.get("ProviderGuid")} processGuid={query.get("ProcessGuid")}/>
    )
}

class Process extends React.Component {
    constructor(props) {
        super(props)
        this.state = {
            procInfo: {}
        }
    }

    componentDidMount() {
        $.ajax({
            url: `${endpoint}?ProviderGuid=${this.props.providerGuid}&ProcessGuid=${this.props.processGuid}`,
            dataType: "json",
            success: function (data) {
                this.setState({
                    procInfo: data,
                })
            }.bind(this),
        })
    }

    render() {
        let proc = this.state.procInfo
        return (
            <div>
                <div>
                    <h3>Execution Details</h3>
                    <p><span>Process ID: {proc.ProcessId}</span></p>
                    <p><span>Image: {proc.Image}</span></p>
                    <p><span>Commandline: {proc.CommandLine}</span></p>
                    <p><span>CurrentDirectory: {proc.CurrentDirectory}</span></p>
                    <p><span>State: {procStates[proc.State]}</span></p>
                    <p><span>Execution time: {proc.CreatedAt}</span></p>
                    <p><span>Stopped At: {proc.TerminatedAt}</span></p>
                    <p><span>Integrity Level: {proc.IntegrityLevel}</span></p>
                </div>

                <div>
                    <h3>File Details</h3>
                    <p><span>SHA1: {proc.SHA1}</span></p>
                    <p><span>SHA256: {proc.SHA256}</span></p>
                    <p><span>MD5: {proc.MD5}</span></p>
                    <p><span>OriginalFileName: {proc.OriginalFileName}</span></p>
                    <p><span>FileVersion: {proc.FileVersion}</span></p>
                    <p><span>Description: {proc.Description}</span></p>
                    <p><span>Product: {proc.Product}</span></p>
                    <p><span>Company: {proc.Company}</span></p>
                </div>
            </div>
        )
    }
}