# fish completion for oci-tf-bootstrap

# Disable file completion by default
complete -c oci-tf-bootstrap -f

# Helper function to get OCI profiles
function __fish_oci_profiles
    if test -f ~/.oci/config
        grep '^\[' ~/.oci/config | tr -d '[]'
    end
end

# OCI regions
set -l regions \
    us-ashburn-1 us-phoenix-1 us-sanjose-1 us-chicago-1 \
    eu-frankfurt-1 eu-amsterdam-1 eu-zurich-1 eu-madrid-1 \
    uk-london-1 uk-cardiff-1 \
    ap-tokyo-1 ap-osaka-1 ap-seoul-1 ap-sydney-1 ap-melbourne-1 ap-mumbai-1 ap-hyderabad-1 \
    ca-toronto-1 ca-montreal-1 \
    sa-saopaulo-1 sa-santiago-1 \
    me-jeddah-1 me-dubai-1 \
    af-johannesburg-1

# Flag completions
complete -c oci-tf-bootstrap -l profile -d 'OCI config profile name' -xa '(__fish_oci_profiles)'
complete -c oci-tf-bootstrap -l config -d 'OCI config directory' -r -a '(__fish_complete_directories)'
complete -c oci-tf-bootstrap -l config-file -d 'OCI config file path' -r -F
complete -c oci-tf-bootstrap -l output -d 'Output directory for generated TF files' -r -a '(__fish_complete_directories)'
complete -c oci-tf-bootstrap -l region -d 'Override region' -xa "$regions"
complete -c oci-tf-bootstrap -l always-free -d 'Filter output to always-free tier eligible resources only'
complete -c oci-tf-bootstrap -l json -d 'Output raw discovery as JSON instead of TF'
complete -c oci-tf-bootstrap -l version -d 'Print version information and exit'
complete -c oci-tf-bootstrap -l help -d 'Show help'
