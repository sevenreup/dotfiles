import QtQuick

// Labelled text input field for the CalDAV account form.
Item {
    id: root

    property string label:       ""
    property string placeholder: ""
    property bool   masked:      false

    implicitHeight: col.implicitHeight
    implicitWidth:  col.implicitWidth

    // Use valueChanged (not textChanged) to avoid conflict with
    // the auto-generated property-change signal from any "text" property.
    signal valueChanged(string val)

    function setText(val) { field.text = val }
    function getValue()   { return field.text }

    Column {
        id: col
        width: parent.width
        spacing: 3

        Text {
            text: root.label
            font.pixelSize: 10
            color: "#6c7086"
        }

        Rectangle {
            width: parent.width
            height: 30
            radius: 6
            color: "#313244"
            border.color: field.activeFocus ? "#cba6f7" : "#45475a"
            border.width: 1

            TextInput {
                id: field
                anchors {
                    left: parent.left; right: parent.right
                    verticalCenter: parent.verticalCenter
                    leftMargin: 8; rightMargin: 8
                }
                echoMode: root.masked ? TextInput.Password : TextInput.Normal
                color: "#cdd6f4"
                font.pixelSize: 12
                selectionColor: "#40cba6f7"
                selectedTextColor: "#cdd6f4"
                clip: true

                Text {
                    anchors.fill: parent
                    text: root.placeholder
                    color: "#45475a"
                    font.pixelSize: 12
                    visible: parent.text === "" && !parent.activeFocus
                    verticalAlignment: Text.AlignVCenter
                }

                onTextChanged: root.valueChanged(text)
            }
        }
    }
}
