Add-Type -AssemblyName System.IO.Compression.FileSystem

Get-ChildItem "Docs\*.docx" | ForEach-Object {
    Write-Output "===================================="
    Write-Output "DOSYA: $($_.Name)"
    Write-Output "===================================="
    $zip = [System.IO.Compression.ZipFile]::OpenRead($_.FullName)
    $entry = $zip.GetEntry("word/document.xml")
    if ($entry) {
        $stream = $entry.Open()
        $reader = New-Object System.IO.StreamReader($stream)
        $xmlContent = $reader.ReadToEnd()
        $reader.Close()
        $stream.Close()
        [xml]$xml = $xmlContent
        $ns = New-Object System.Xml.XmlNamespaceManager($xml.NameTable)
        $ns.AddNamespace("w", "http://schemas.openxmlformats.org/wordprocessingml/2006/main")
        $nodes = $xml.SelectNodes("//w:p", $ns)
        foreach ($node in $nodes) {
            $texts = $node.SelectNodes(".//w:t", $ns)
            $line = ""
            foreach ($t in $texts) {
                $line += $t.InnerText
            }
            if ($line.Trim().Length -gt 0) {
                Write-Output $line
            }
        }
    }
    $zip.Dispose()
}
