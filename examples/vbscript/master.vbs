Sub Master()
    Dim aimasteringPath
    Dim command
    Dim shell
    Dim exitCode

    aimasteringPath = "C:\aimastering-windows-386.exe"
    command = Array(aimasteringPath, "master", "--input", "C:\input.wav", "--output", "C:\output.wav")
    command = """" & Join(command, """ """) & """"
    WScript.Echo command

    Set shell = WScript.CreateObject("WScript.Shell")
    exitCode = shell.Run(command, 1, True)
    If exitCode = 0 Then
      WScript.Echo "Succeeded"
    Else
      WScript.Echo "Failed"
    End If
End Sub

Call Master()
