#compdef oci-tf-bootstrap

# zsh completion for oci-tf-bootstrap

_oci_tf_bootstrap() {
    local -a opts
    local -a regions

    regions=(
        'us-ashburn-1:US East (Ashburn)'
        'us-phoenix-1:US West (Phoenix)'
        'us-sanjose-1:US West (San Jose)'
        'us-chicago-1:US Midwest (Chicago)'
        'eu-frankfurt-1:Germany Central (Frankfurt)'
        'eu-amsterdam-1:Netherlands Northwest (Amsterdam)'
        'eu-zurich-1:Switzerland North (Zurich)'
        'eu-madrid-1:Spain Central (Madrid)'
        'uk-london-1:UK South (London)'
        'uk-cardiff-1:UK West (Cardiff)'
        'ap-tokyo-1:Japan East (Tokyo)'
        'ap-osaka-1:Japan Central (Osaka)'
        'ap-seoul-1:South Korea Central (Seoul)'
        'ap-sydney-1:Australia East (Sydney)'
        'ap-melbourne-1:Australia Southeast (Melbourne)'
        'ap-mumbai-1:India West (Mumbai)'
        'ap-hyderabad-1:India South (Hyderabad)'
        'ca-toronto-1:Canada Southeast (Toronto)'
        'ca-montreal-1:Canada Southeast (Montreal)'
        'sa-saopaulo-1:Brazil East (Sao Paulo)'
        'sa-santiago-1:Chile Central (Santiago)'
        'me-jeddah-1:Saudi Arabia West (Jeddah)'
        'me-dubai-1:UAE East (Dubai)'
        'af-johannesburg-1:South Africa Central (Johannesburg)'
    )

    _arguments -s \
        '--profile[OCI config profile name]:profile:->profiles' \
        '--config[OCI config directory]:directory:_files -/' \
        '--config-file[OCI config file path]:file:_files' \
        '--output[Output directory for generated TF files]:directory:_files -/' \
        '--region[Override region]:region:->regions' \
        '--always-free[Filter output to always-free tier eligible resources only]' \
        '--json[Output raw discovery as JSON instead of TF]' \
        '--version[Print version information and exit]' \
        '--help[Show help]'

    case "$state" in
        profiles)
            local -a profiles
            if [[ -f ~/.oci/config ]]; then
                profiles=(${(f)"$(grep '^\[' ~/.oci/config | tr -d '[]')"})
            fi
            _describe -t profiles 'OCI profile' profiles
            ;;
        regions)
            _describe -t regions 'OCI region' regions
            ;;
    esac
}

_oci_tf_bootstrap "$@"
